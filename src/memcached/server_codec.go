package memcached

import (
	"errors"
	"fmt"
	"io"
	. "logger"
	. "sock"
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
	mc.buffer = NewBuffer(1024, -1)
	mc.reader = reader
	mc.writer = writer
	return mc
}

type MemcachedServerCodec struct {
	buffer *Buffer
	reader io.Reader
	writer io.Writer
}

func (c *MemcachedServerCodec) Decode() (interface{}, error) {
	var from, pos int = -1, -1
	pos = c.buffer.FindCRLF(from)
	for pos < 0 { // need more data
		n, err := c.buffer.ReadFrom(c.reader)
		LOG.Debug("pos = %d, n = %d, err = %v", pos, n, err)

		if n > 0 {
			pos = c.buffer.FindCRLF(from)
			if c.buffer.Len() > MAX_LINE_LENGTH && pos < 0 {
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
	_, err := c.buffer.Read(line)
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
	return h(cmd, c.buffer, c.reader, tk)
}

func (c *MemcachedServerCodec) Encode(body interface{}) error {
	buffers := make([][]byte, 0, 8)
	resp, _ := body.(*MemResponse)
	if "" != resp.Err {
		LOG.Debug("[Encode] Op = %d, err = %s\n", resp.Op, resp.Err)
		buffers = append(buffers, []byte(resp.Err))
	} else if resp.Op == SET || resp.Op == ADD || resp.Op == REPLACE {
		if resp.Result {
			buffers = append(buffers, []byte("STORED\r\n"))
		} else {
			buffers = append(buffers, []byte("NOT_STORED\r\n"))
		}
	} else if resp.Op == DELETE {
		if resp.Result {
			buffers = append(buffers, []byte("DELETED\r\n"))
		} else {
			buffers = append(buffers, []byte("NOT_FOUND\r\n"))
		}
	} else if resp.Op == GET {
		if resp.Result {
			line := fmt.Sprintf("VALUE %s %d %d\r\n", resp.Key, resp.Flags, resp.Bytes)

			buffers = append(buffers, []byte(line))
			buffers = append(buffers, resp.Data)
			buffers = append(buffers, []byte("\r\nEND\r\n"))
		} else {
			buffers = append(buffers, []byte("END\r\n"))
		}
	} else if resp.Op == INCR || resp.Op == DECR {
		if resp.Result {
			buffers = append(buffers, []byte(fmt.Sprintf("%d\r\n", resp.Value)))
		} else {
			buffers = append(buffers, []byte("NOT_FOUND\r\n"))
		}
	} else {
		panic("unknow cmd")
		//return &MemRequest{Err: "ERROR OP\r\n"}, nil
	}
	LOG.Debug("[Encode] '%v'\n", buffers)
	for _, buffer := range buffers {
		_, err := c.writer.Write(buffer)
		if err != nil {
			return err
		}
	}
	return nil
}
