package memcached

import (
	"bytes"
	"errors"
	. "github.com/elwin0214/gomemcached/sock"
	"github.com/golang/glog"
	"io"
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
	VER     = Code(8)
	QUIT    = Code(9)
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
	"ver":     handleVersionRequest,
}
var cmds = map[Code][]byte{
	SET:     []byte("set"),
	ADD:     []byte("add"),
	REPLACE: []byte("replace"),
	GET:     []byte("get"),
	DELETE:  []byte("delete"),
	INCR:    []byte("incr"),
	DECR:    []byte("decr"),
	VER:     []byte("version"),
}

func NewMemcachedServerCodec(reader io.Reader, writer io.Writer, readBufferSize int) Codec {
	mc := new(MemcachedServerCodec)
	mc.rb = NewBuffer(readBufferSize, -1)
	mc.reader = reader
	mc.writer = writer
	return mc
}

type MemcachedServerCodec struct {
	rb     *Buffer
	reader io.Reader
	writer io.Writer
}

func (c *MemcachedServerCodec) Decode() (interface{}, error) {
	var from, pos int = -1, -1
	pos = c.rb.FindCRLF(from)
	for pos < 0 {
		// need more data
		n, err := c.rb.ReadFrom(c.reader)
		glog.Infof("[Decode] pos = %d, n = %d, err = %v", pos, n, err)

		if n > 0 {
			pos = c.rb.FindCRLF(from)
			if c.rb.Len() > MAX_LINE_LENGTH && pos < 0 {
				return nil, ErrTooLongLine
			}
		}
		glog.Infof("pos = %d, n = %d, err = %v", pos, n, err)
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
	if cmd == "quit" {
		return nil, errors.New("Connection goto close")
	}
	var h handler

	h, ok = handlers[cmd]
	if !ok {
		glog.Infof("[Decode] cmd = %s\n", cmd)
		return &MemRequest{Err: "ERROR\r\n"}, nil
	}
	return h(cmd, c.rb, c.reader, tk)
}

func (c *MemcachedServerCodec) Encode(body interface{}) error {
	resp, _ := body.(*MemResponse)
	//todo write error
	var buf []byte

	if "" != resp.Err {
		glog.Infof("[Encode] Op = %d, err = %s\n", resp.Op, resp.Err)
		buf = []byte(resp.Err)
	} else if resp.Op == SET || resp.Op == ADD || resp.Op == REPLACE {
		if resp.Result {
			buf = []byte("STORED\r\n")
		} else {
			buf = []byte("NOT_STORED\r\n")
		}
	} else if resp.Op == DELETE {
		if resp.Result {
			buf = []byte("DELETED\r\n")
		} else {
			buf = []byte("NOT_FOUND\r\n")
		}
	} else if resp.Op == INCR || resp.Op == DECR {
		buffer := bytes.NewBuffer(make([]byte, 0, 1024))
		if resp.Result {
			buffer.WriteString(strconv.FormatInt(int64(resp.Value), 10))
			buffer.WriteString("\r\n")
		} else {
			buffer.Write([]byte("NOT_FOUND\r\n"))
		}
		buf = buffer.Bytes()
	} else if resp.Op == GET {
		if resp.Result {
			buffer := bytes.NewBuffer(make([]byte, 0, len(resp.Data)+1024))
			buffer.WriteString("VALUE ")
			buffer.WriteString(resp.Key)
			buffer.WriteString(" ")
			buffer.WriteString(strconv.FormatInt(int64(resp.Flags), 10))
			buffer.WriteString(" ")
			buffer.WriteString(strconv.FormatInt(int64(resp.Bytes), 10))
			buffer.WriteString("\r\n")
			buffer.Write(resp.Data)
			buffer.WriteString("\r\nEND\r\n")
			buf = buffer.Bytes()
		} else {
			buf = []byte("END\r\n")
		}
	} else {
		panic("unknow cmd")
	}
	_, err := c.writer.Write(buf)
	return err
}
