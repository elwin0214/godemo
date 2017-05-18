package memcached

import (
	"bytes"
	"errors"
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
	mc.rb = NewBuffer(4096, -1)
	mc.wb = bytes.NewBuffer(make([]byte, 0, 4096))

	mc.reader = reader
	mc.writer = writer
	return mc
}

type MemcachedClientCodec struct {
	rb     *Buffer
	wb     *bytes.Buffer
	reader io.Reader
	writer io.Writer
}

func (c *MemcachedClientCodec) Decode() (interface{}, error) {
	var from, pos int = -1, -1
	pos = c.rb.FindCRLF(from)
	for pos < 0 { // need more data
		n, err := c.rb.ReadFrom(c.reader)
		if n > 0 {
			pos = c.rb.FindCRLF(from)
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
	_, err := c.rb.Read(buf)
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
		for !(c.rb.Len() >= int(resp.Bytes)+2 && c.rb.FindCRLF(int(resp.Bytes)) > 0) { // zero length?
			n, err := c.rb.ReadFrom(c.reader)
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
		c.rb.Read(resp.Data)
		c.rb.Skip(2)
		pos = c.rb.FindCRLF(0)
		for pos < 0 { // need more data
			n, err := c.rb.ReadFrom(c.reader)
			if n > 0 {
				pos = c.rb.FindCRLF(0)
			}
			if nil != err {
				LOG.Error("[Decode] error = %s\n", err.Error())
				return nil, err
			}
			if n == 0 {
				return nil, errors.New("Connection closed")
			}
		}
		c.rb.Skip(pos + 2)
		LOG.Debug("[Decode] 2 buf = %s\n", string(c.rb.Bytes()))
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
		//too large panic
		c.wb.Write(cmds[r.Op])
		c.wb.WriteString(" ")
		c.wb.WriteString(r.Key)
		c.wb.WriteString(" ")
		c.wb.WriteString(strconv.FormatInt(int64(r.Flags), 10))
		c.wb.WriteString(" ")
		c.wb.WriteString(strconv.FormatInt(int64(r.Exptime), 10))
		c.wb.WriteString(" ")
		c.wb.WriteString(strconv.FormatInt(int64(r.Bytes), 10))
		c.wb.WriteString("\r\n")
		c.wb.Write(r.Data)
		c.wb.WriteString("\r\n")
	}

	if r.Op == DELETE {
		c.wb.Write(cmds[r.Op])
		c.wb.WriteString(" ")
		c.wb.WriteString(r.Key)
		c.wb.WriteString("\r\n")
	}
	if r.Op == GET {
		c.wb.Write(cmds[r.Op])
		c.wb.WriteString(" ")
		c.wb.WriteString(r.Key)
		c.wb.WriteString("\r\n")
	}
	if r.Op == INCR || r.Op == DECR {
		c.wb.Write(cmds[r.Op])
		c.wb.WriteString(" ")
		c.wb.WriteString(r.Key)
		c.wb.WriteString(" ")
		c.wb.WriteString(strconv.FormatInt(int64(r.Value), 10))
		c.wb.WriteString("\r\n")
	}
	_, err := c.writer.Write(c.wb.Bytes())
	c.wb.Reset()
	if nil != err {
		return err
	}
	return nil
}
