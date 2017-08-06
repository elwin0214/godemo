package memcached

import (
	. "memcached"
	"testing"
)

type StreamReader struct {
	data []byte
}

func newStreamReader() *StreamReader {
	return &StreamReader{data: make([]byte, 0, 1024)}
}

func (s *StreamReader) Read(buf []byte) (int, error) {
	size := len(buf)
	if size >= len(s.data) {
		size = len(s.data)
	}
	copy(buf, s.data[0:size])
	if size == len(s.data) {
		s.data = s.data[0:0]
	} else {
		copy(s.data[0:len(s.data)-size], s.data[size:])
		s.data = s.data[0 : len(s.data)-size]
	}
	return size, nil
}

func (s *StreamReader) append(buf []byte) {
	s.data = append(s.data, buf...)
}

func Test_boundary(t *testing.T) {
	t.Logf("%d\n", MAX_LINE_LENGTH)
	t.Logf("%d\n", MAX_KEY_LENGTH)
	t.Logf("%d\n", MAX_UINT32_VALUE)
	t.Logf("%d\n", MAX_UINT16_VALUE)
}

func Test_Reader(t *testing.T) {
	reader := newStreamReader()
	reader.append([]byte("abc"))
	buf := make([]byte, 1024, 1024)
	n, _ := reader.Read(buf)
	t.Logf("%d\n", n)
	if "abc" != string(buf[0:n]) {
		t.Errorf("buf is '%s'\n", string(buf))
	}
	reader.append([]byte("defghk"))
	n, _ = reader.Read(buf)
	if "defghk" != string(buf[0:n]) {
		t.Errorf("buf is '%s'\n", string(buf))
	}
}

func Test_Pack(t *testing.T) {
	reader := newStreamReader()
	line := "set a 0 1 1\r\na\r\nset a 0 1 "
	reader.append([]byte(line)) //1\r\na\r\n
	coder := NewMemcachedServerCodec(reader, nil, 1024)
	req, err := coder.Decode()
	if nil != err {
		t.Errorf("%s for '%s'\n", err.Error(), line)
	}
	t.Log(req)
	if r, _ := req.(*MemRequest); SET != r.Op {
		t.Errorf("parse '%s' fail\n", line)
	}
	line = "1\r\na\r\n"
	reader.append([]byte(line))
	req, err = coder.Decode()
	if nil != err {
		t.Errorf("%s for '%s'\n", err.Error(), line)
	}

	if r, _ := req.(*MemRequest); SET != r.Op {
		t.Errorf("parse '%s' fail\n", line)
	}
}
