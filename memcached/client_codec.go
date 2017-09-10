package memcached

import (
	"bytes"
	"errors"
	. "github.com/elwin0214/gomemcached/sock"
	"github.com/golang/glog"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var NUMBER_REG = regexp.MustCompile("^[0-9]+$")

const (
	defaultBufSize = 32 * 1024
)

func NewMemcachedClientCodec(reader io.Reader, writer io.Writer) Codec {
	mc := new(MemcachedClientCodec)
	mc.rb = NewBuffer(defaultBufSize, -1)
	mc.reader = reader
	mc.writer = writer
	return mc
}

type MemcachedClientCodec struct {
	rb     *Buffer
	reader io.Reader
	writer io.Writer
}

func (c *MemcachedClientCodec) Decode() (interface{}, error) {
	var from, pos int = -1, -1
	pos = c.rb.FindCRLF(from)
	for pos < 0 {
		// need more data
		n, err := c.rb.ReadFrom(c.reader)
		if n > 0 {
			pos = c.rb.FindCRLF(from)
		}
		if nil != err {
			glog.Errorf("[Decode] error = %s\n", err.Error())
			return nil, err
		}
		if n == 0 {
			glog.Errorf("[Decode] goto close conn\n")
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
		for !(c.rb.Len() >= int(resp.Bytes)+2 && c.rb.FindCRLF(int(resp.Bytes)) > 0) {
			// zero length?
			n, err := c.rb.ReadFrom(c.reader)
			if n > 0 {
				continue
			}
			if nil != err {
				glog.Errorf("[Decode] error = %s\n", err.Error())
				return nil, err
			}
			if n == 0 {
				glog.Errorf("[Decode] goto close conn\n")
				return nil, errors.New("Connection closed")
			}
		}

		resp.Data = make([]byte, resp.Bytes, resp.Bytes)
		c.rb.Read(resp.Data)
		c.rb.Skip(2)
		pos = c.rb.FindCRLF(0)
		for pos < 0 {
			// need more data
			n, err := c.rb.ReadFrom(c.reader)
			if n > 0 {
				pos = c.rb.FindCRLF(0)
			}
			if nil != err {
				glog.Errorf("[Decode] error = %s\n", err.Error())
				return nil, err
			}
			if n == 0 {
				return nil, errors.New("Connection closed")
			}
		}
		c.rb.Skip(pos + 2)
		glog.Infof("[Decode] 2 buf = %s\n", string(c.rb.Bytes()))
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
	var buf []byte
	if r.Op == SET || r.Op == ADD || r.Op == REPLACE {
		//too large panic
		//todo writer error
		buffer := bytes.NewBuffer(make([]byte, 0, len(r.Data)+1024))
		buffer.Write(cmds[r.Op])
		buffer.WriteString(" ")
		buffer.WriteString(r.Key)
		buffer.WriteString(" ")
		buffer.WriteString(strconv.FormatInt(int64(r.Flags), 10))
		buffer.WriteString(" ")
		buffer.WriteString(strconv.FormatInt(int64(r.Exptime), 10))
		buffer.WriteString(" ")
		buffer.WriteString(strconv.FormatInt(int64(r.Bytes), 10))
		buffer.WriteString("\r\n")
		buffer.Write(r.Data)
		buffer.WriteString("\r\n")
		buf = buffer.Bytes()
	}

	if r.Op == DELETE {
		buffer := bytes.NewBuffer(make([]byte, 0, 1024))
		buffer.Write(cmds[r.Op])
		buffer.WriteString(" ")
		buffer.WriteString(r.Key)
		buffer.WriteString("\r\n")
		buf = buffer.Bytes()
	}
	if r.Op == GET {
		buffer := bytes.NewBuffer(make([]byte, 0, 1024))
		buffer.Write(cmds[r.Op])
		buffer.WriteString(" ")
		buffer.WriteString(r.Key)
		buffer.WriteString("\r\n")
		buf = buffer.Bytes()
	}
	if r.Op == INCR || r.Op == DECR {
		buffer := bytes.NewBuffer(make([]byte, 0, 1024))
		buffer.Write(cmds[r.Op])
		buffer.WriteString(" ")
		buffer.WriteString(r.Key)
		buffer.WriteString(" ")
		buffer.WriteString(strconv.FormatInt(int64(r.Value), 10))
		buffer.WriteString("\r\n")
		buf = buffer.Bytes()
	}

	if r.Op == VER {
		buffer := bytes.NewBuffer(make([]byte, 0, 16))
		buffer.Write(cmds[r.Op])
		buffer.WriteString("\r\n")
		buf = buffer.Bytes()
	}
	_, err := c.writer.Write(buf)
	return err
}
