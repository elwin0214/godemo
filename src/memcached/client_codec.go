package memcached

import (
	"errors"
	"fmt"
	"io"
	. "logger"
	"regexp"
	. "sock"
	"strconv"
	"strings"
)

var NUMBER_REG = regexp.MustCompile("^[0-9]+$")

func NewMemcachedClientCodec(reader io.Reader, writer io.Writer) Codec {
	mc := new(MemcachedClientCodec)
	mc.buffer = NewBuffer(1024, -1)
	mc.reader = reader
	mc.writer = writer
	return mc
}

type MemcachedClientCodec struct {
	buffer *Buffer
	reader io.Reader
	writer io.Writer
}

func (c *MemcachedClientCodec) Decode() (interface{}, error) {
	var from, pos int = -1, -1
	pos = c.buffer.FindCRLF(from)
	for pos < 0 { // need more data
		n, err := c.buffer.ReadFrom(c.reader)
		if n > 0 {
			pos = c.buffer.FindCRLF(from)
		}
		if nil != err {
			LOG.Error("[Decode] error = %s\n", err.Error())
			return nil, err
		}
		if n == 0 {
			LOG.Error("[Decode] goto close conn\n")
			return nil, errors.New("Connection closed")
		}
	}

	buf := make([]byte, pos+2, pos+2)
	_, err := c.buffer.Read(buf)
	if nil != err {
		return nil, err
	}
	buf = buf[0:pos]

	tk := NewTokenizer(buf, ' ')
	ok, bytes := tk.Next()
	if !ok {
		return &MemResponse{Result: false, Err: "ERROR"}, nil
	}

	word := string(bytes)
	resp := new(MemResponse)

	if "VALUE" == word {
		ok, word := tk.Next()
		if !ok {
			resp.Err = "ERROR"
			resp.Result = false
			return resp, nil
		}
		resp.Key = string(word)

		ok, word = tk.Next()
		if !ok {
			resp.Err = "ERROR"
			resp.Result = false
			return resp, nil
		}
		var flag int
		flag, err = (strconv.Atoi(string(word)))
		if nil != err {
			resp.Err = "ERROR"
			resp.Result = false
			return resp, nil
		}
		resp.Flags = uint32(flag)

		ok, word = tk.Next()
		if !ok {
			resp.Err = "ERROR"
			resp.Result = false
			return resp, nil
		}
		//todo overflow
		var bytes int
		bytes, err = (strconv.Atoi(string(word)))
		if nil != err {
			resp.Err = "ERROR"
			resp.Result = false
			return resp, nil
		}
		resp.Bytes = uint16(bytes)
		//LOG.Info("buf.Len() = %d bytes = %d", buf.Len(), resp.Bytes)
		for !(c.buffer.Len() >= int(resp.Bytes)+2 && c.buffer.FindCRLF(int(resp.Bytes)) > 0) { // zero length?
			n, err := c.buffer.ReadFrom(c.reader)
			if n > 0 {
				continue
			}
			if nil != err {
				LOG.Error("[Decode] error = %s\n", err.Error())
				return nil, err
			}
			if n == 0 {
				LOG.Error("[Decode] goto close conn\n")
				return nil, errors.New("Connection closed")
			}
		}

		resp.Data = make([]byte, resp.Bytes, resp.Bytes)
		c.buffer.Read(resp.Data)
		c.buffer.Skip(2)
		pos = c.buffer.FindCRLF(0)
		for pos < 0 { // need more data
			n, err := c.buffer.ReadFrom(c.reader)
			if n > 0 {
				pos = c.buffer.FindCRLF(0)
			}
			if nil != err {
				LOG.Error("[Decode] error = %s\n", err.Error())
				return nil, err
			}
			if n == 0 {
				return nil, errors.New("Connection closed")
			}
		}
		c.buffer.Skip(pos + 2)
		LOG.Debug("[Decode] 2 buf = %s\n", string(c.buffer.Bytes()))
		resp.Result = true
		return resp, nil
	}

	line := string(buf)
	if strings.HasPrefix(line, "STORED") {
		return &MemResponse{Result: true}, nil
	}
	if strings.HasPrefix(line, "NOT_STORED") {
		return &MemResponse{Result: false, Err: line}, nil
	}
	if strings.HasPrefix(line, "DELETED") {
		return &MemResponse{Result: true}, nil
	}
	if strings.HasPrefix(line, "NOT_FOUND") {
		return &MemResponse{Result: false, Err: line}, nil
	}
	if strings.HasPrefix(line, "ERROR") {
		return &MemResponse{Result: false, Err: line}, nil
	}
	if strings.HasPrefix(line, "ERROR") {

	}
	if !NUMBER_REG.MatchString(line) {
		return &MemResponse{Result: false, Err: line}, nil
	} else {
		value, e := strconv.ParseInt(string(word), 10, 64)
		if nil == e {
			return &MemResponse{Result: true, Value: uint32(value)}, nil
		} else {
			return &MemResponse{Result: false, Err: line}, nil
		}
	}
}

func (c *MemcachedClientCodec) Encode(req interface{}) error {
	r, _ := req.(*MemRequest)
	if r.Op == SET || r.Op == ADD || r.Op == REPLACE {
		line := fmt.Sprintf("%s %s %d %d %d\r\n", cmds[r.Op], r.Key, r.Flags, r.Exptime, r.Bytes)
		_, err := c.writer.Write([]byte(line))
		if nil != err {
			return err
		}
		_, err = c.writer.Write(r.Data)
		if nil != err {
			return err
		}
		_, err = c.writer.Write([]byte("\r\n"))
		if nil != err {
			return err
		}
	}

	if r.Op == DELETE {
		line := fmt.Sprintf("%s %s\r\n", cmds[r.Op], r.Key)
		_, err := c.writer.Write([]byte(line))
		if nil != err {
			return err
		}
	}
	if r.Op == GET {
		line := fmt.Sprintf("%s %s\r\n", cmds[r.Op], r.Key)
		_, err := c.writer.Write([]byte(line))
		if nil != err {
			return err
		}
	}
	if r.Op == INCR || r.Op == DECR {
		line := fmt.Sprintf("%s %s %d\r\n", cmds[r.Op], r.Key, r.Value)
		_, err := c.writer.Write([]byte(line))
		if nil != err {
			return err
		}
	}
	return nil
}
