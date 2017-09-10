package memcached

import (
	"errors"
	. "github.com/elwin0214/gomemcached/sock"
	"github.com/golang/glog"
	"io"
	"strconv"
)

var ErrTooLongKey = errors.New("key: too long")
var ErrTooLongData = errors.New("data: too long")

func handleStoreRequest(cmd string, buf *Buffer, reader io.Reader, tk *Tokenizer) (*MemRequest, error) {
	req := new(MemRequest)
	if "set" == cmd {
		req.Op = SET
	} else if "add" == cmd {
		req.Op = ADD
	} else {
		req.Op = REPLACE
	}
	ok, word := tk.Next()
	if !ok {
		//fmt.Sprintf(format, ...)
		return &MemRequest{Err: "ERROR KEY\r\n"}, nil
	}
	if len(word) > (MAX_KEY_LENGTH) {
		return nil, ErrTooLongKey
	}
	req.Key = string(word)
	ok, word = tk.Next()
	if !ok {
		return &MemRequest{Err: "ERROR FLAG\r\n"}, nil
	}
	var flag int64
	var err error
	flag, err = strconv.ParseInt(string(word), 10, 64)
	if nil != err {
		return &MemRequest{Err: "ERROR FLAG\r\n"}, nil
	}
	if flag > MAX_UINT32_VALUE {
		return &MemRequest{Err: "ERROR FLAG OVERFLOW\r\n"}, nil
	}
	req.Flags = uint32(flag)
	ok, word = tk.Next()
	if !ok {
		return &MemRequest{Err: "ERROR EXPTIME\r\n"}, nil
	}
	//todo overflow
	var exptime int64
	exptime, err = strconv.ParseInt(string(word), 10, 64)
	if nil != err {
		return &MemRequest{Err: "ERROR EXPTIME\r\n"}, nil
	}
	if exptime > MAX_UINT32_VALUE {
		return &MemRequest{Err: "ERROR EXPTIME OVERFLOW\r\n"}, nil
	}
	req.Exptime = uint32(exptime)
	ok, word = tk.Next()
	if !ok {
		return &MemRequest{Err: "ERROR BYTES\r\n"}, nil
	}
	//todo overflow
	var bytes int
	bytes, err = strconv.Atoi(string(word))
	if nil != err {
		return &MemRequest{Err: "ERROR\r\n"}, nil
	}
	if bytes > MAX_UINT16_VALUE {
		return &MemRequest{Err: "ERROR BYTES OVERFLOW\r\n"}, nil
	}
	req.Bytes = uint16(bytes)
	glog.Infof("buf.Len() = %d bytes = %d", buf.Len(), req.Bytes)
	for { // zero length?
		hitLen := buf.Len() >= int(req.Bytes)+2
		pos := buf.FindCRLF(int(req.Bytes))
		hitCrlf := pos > 0
		if hitLen && hitCrlf {
			if int(req.Bytes) < pos {
				buf.Skip(pos + 2)
				return &MemRequest{Err: "ERROR DATA\r\n"}, nil
			}
			break
		}
		if hitLen && !hitCrlf && buf.Len()+2 > MAX_UINT16_VALUE {
			return nil, ErrTooLongData
		}
		n, err := buf.ReadFrom(reader)

		if n > 0 {
			continue
		}
		if nil != err {
			return nil, err
		}
		if n == 0 {
			glog.Infof("%d, %v\n", n, err)
			return nil, errors.New("Connection closed")
		}
	}

	req.Data = make([]byte, req.Bytes, req.Bytes)
	buf.Read(req.Data)
	buf.Skip(2)
	return req, nil
}

func handleDeleteRequest(cmd string, buf *Buffer, reader io.Reader, tk *Tokenizer) (r *MemRequest, err error) {
	req := new(MemRequest)
	req.Op = DELETE
	ok, word := tk.Next()
	if !ok {
		return &MemRequest{Err: "ERROR"}, nil
	}
	if len(word) > MAX_KEY_LENGTH {
		return nil, ErrTooLongKey
	}
	req.Key = string(word)
	return req, nil
}
func handleGetRequest(cmd string, buf *Buffer, reader io.Reader, tk *Tokenizer) (r *MemRequest, err error) {
	req := new(MemRequest)
	req.Op = GET
	ok, word := tk.Next()
	if !ok {
		return &MemRequest{Err: "ERROR"}, nil
	}
	if len(word) > MAX_KEY_LENGTH {
		return nil, ErrTooLongKey
	}
	req.Key = string(word)
	return req, nil
}

func handleCounterRequest(cmd string, buf *Buffer, reader io.Reader, tk *Tokenizer) (*MemRequest, error) {
	req := new(MemRequest)

	if "incr" == cmd {
		req.Op = INCR
	} else {
		req.Op = DECR
	}
	ok, word := tk.Next()
	if !ok {
		return &MemRequest{Err: "ERROR"}, nil
	}
	req.Key = string(word)
	ok, word = tk.Next()
	if !ok {
		return &MemRequest{Err: "ERROR"}, nil
	}
	if len(word) > MAX_KEY_LENGTH {
		return nil, ErrTooLongKey
	}
	// todo overflow
	var value int64
	var err error
	value, err = strconv.ParseInt(string(word), 10, 64)
	if nil != err {
		return &MemRequest{Err: "ERROR"}, nil
	}
	if value > MAX_UINT32_VALUE {
		return &MemRequest{Err: "ERROR VALUE OVERFLOW"}, nil
	}
	req.Value = uint32(value)
	return req, nil
}

func handleVersionRequest(cmd string, buf *Buffer, reader io.Reader, tk *Tokenizer) (*MemRequest, error) {
	return &MemRequest{Err: "V1.0"}, nil
}
