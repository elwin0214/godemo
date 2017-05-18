package memcached

import (
	"bytes"
	"errors"
	"io"
	. "logger"
	. "sock"
	"strconv"
)

type Code uint8

const (
	SET     = Code(1)
	ADD     = Code(2)
	REPLACE = Code(3)
	GET     = Code(4)
	DELETE  = Code(5)
	INCR    = Code(6)
	DECR    = Code(7)
)

const (
	MAX_KEY_LENGTH   = int(1<<8 - 1)
	MAX_LINE_LENGTH  = int(1<<16 - 1)
	MAX_UINT32_VALUE = int64(1<<32 - 1)
	MAX_UINT16_VALUE = int(1<<16 - 1)
)

var ErrTooLongLine = errors.New("command line: too long")

type handler func(cmd string, buf *Buffer, reader io.Reader, tk *Tokenizer) (r *MemRequest, err error)

var handlers = map[string]handler{
	"set":     handleStoreRequest,
	"add":     handleStoreRequest,
	"replace": handleStoreRequest,
	"delete":  handleDeleteRequest,
	"get":     handleGetRequest,
	"incr":    handleCounterRequest,
	"decr":    handleCounterRequest,
}
var cmds = map[Code][]byte{
	SET:     []byte("set"),
	ADD:     []byte("add"),
	REPLACE: []byte("replace"),
	GET:     []byte("get"),
	DELETE:  []byte("delete"),
	INCR:    []byte("incr"),
	DECR:    []byte("decr"),
}

func NewMemcachedServerCodec(reader io.Reader, writer io.Writer) Codec {
	mc := new(MemcachedServerCodec)
	mc.rb = NewBuffer(4096, -1)
	mc.wb = bytes.NewBuffer(make([]byte, 0, 4096))

	mc.reader = reader
	mc.writer = writer
	return mc
}

type MemcachedServerCodec struct {
	rb     *Buffer
	wb     *bytes.Buffer
	reader io.Reader
	writer io.Writer
}

func (c *MemcachedServerCodec) Decode() (interface{}, error) {
	var from, pos int = -1, -1
	pos = c.rb.FindCRLF(from)
	for pos < 0 { // need more data
		n, err := c.rb.ReadFrom(c.reader)
		LOG.Debug("pos = %d, n = %d, err = %v", pos, n, err)

		if n > 0 {
			pos = c.rb.FindCRLF(from)
			if c.rb.Len() > MAX_LINE_LENGTH && pos < 0 {
				return nil, ErrTooLongLine
			}
		}
		LOG.Debug("pos = %d, n = %d, err = %v", pos, n, err)
		if nil != err {
			return nil, err
		}
		if n == 0 {
			return nil, errors.New("Connection closed")
		}
	}
	line := make([]byte, pos+2, pos+2)
	_, err := c.rb.Read(line)
	if nil != err {
		return nil, err
	}
	line = line[0:pos]

	tk := NewTokenizer(line, ' ')
	ok, word := tk.Next()
	if !ok {
		return &MemRequest{Err: "ERROR OP\r\n"}, nil
	}
	cmd := string(word)
	if cmd == "exit" {
		return nil, errors.New("Connection goto close")
	}
	var h handler

	h, ok = handlers[cmd]
	if !ok {
		LOG.Debug("[Decode] cmd = %s\n", cmd)
		return &MemRequest{Err: "ERROR\r\n"}, nil
	}
	return h(cmd, c.rb, c.reader, tk)
}

func (c *MemcachedServerCodec) Encode(body interface{}) error {
	resp, _ := body.(*MemResponse)

	var buffer []byte
	if "" != resp.Err {
		LOG.Debug("[Encode] Op = %d, err = %s\n", resp.Op, resp.Err)
		buffer = []byte(resp.Err)
	} else if resp.Op == SET || resp.Op == ADD || resp.Op == REPLACE {
		if resp.Result {
			buffer = []byte("STORED\r\n")
		} else {
			buffer = []byte("NOT_STORED\r\n")
		}
	} else if resp.Op == DELETE {
		if resp.Result {
			buffer = []byte("DELETED\r\n")
		} else {
			buffer = []byte("NOT_FOUND\r\n")
		}
	} else if resp.Op == INCR || resp.Op == DECR {
		if resp.Result {
			c.wb.WriteString(strconv.FormatInt(int64(resp.Value), 10))
			c.wb.WriteString("\r\n")
			buffer = c.wb.Bytes()
		} else {
			buffer = []byte("NOT_FOUND\r\n")
		}
	} else if resp.Op == GET {
		if resp.Result {
			c.wb.WriteString("VALUE ")
			c.wb.WriteString(resp.Key)
			c.wb.WriteString(" ")
			c.wb.WriteString(strconv.FormatInt(int64(resp.Flags), 10))
			c.wb.WriteString(" ")
			c.wb.WriteString(strconv.FormatInt(int64(resp.Bytes), 10))
			c.wb.WriteString("\r\n")
			c.wb.Write(resp.Data)
			c.wb.WriteString("\r\nEND\r\n")
			buffer = c.wb.Bytes()
		} else {
			buffer = []byte("END\r\n")
		}
	} else {
		panic("unknow cmd")

	}
	// if len(buffer) == 0 {

	// }
	LOG.Debug("[Encode] '%v'\n", buffer)
	_, err := c.writer.Write(buffer)
	c.wb.Reset()
	if err != nil {
		return err
	}
	return nil
}
