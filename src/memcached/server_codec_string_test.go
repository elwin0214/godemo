package memcached

import (
	. "memcached"
	"strings"
	"testing"
)

func Test_Store(t *testing.T) {
	line := "set"
	reader := strings.NewReader(line)
	coder := NewMemcachedServerCodec(reader, nil)
	_, err := coder.Decode()
	if nil == err {
		t.Errorf("can not read EOF for '%s'\n", line)
	}
}

func Test_Set(t *testing.T) {
	line := "set a 0 1 1\r\na\r\n"
	reader := strings.NewReader(line)
	coder := NewMemcachedServerCodec(reader, nil)
	req, err := coder.Decode()
	if nil != err {
		t.Errorf("%s for '%s'\n", err.Error(), line)
	}

	if r, _ := req.(*MemRequest); SET != r.Op {
		t.Errorf("parse '%s' fail\n", line)
	}
	req, err = coder.Decode()
	if nil == err {
		t.Errorf("can not get EOF\n")
	}
}

func Test_Set_EOF(t *testing.T) {
	line := "set a 0 1 1\r\na\r"
	reader := strings.NewReader(line)
	coder := NewMemcachedServerCodec(reader, nil)
	_, err := coder.Decode()
	if nil == err {
		t.Errorf("can not get EOF\n")
	}
}

func Test_Get(t *testing.T) {
	line := "get a\r\n"
	reader := strings.NewReader(line)
	coder := NewMemcachedServerCodec(reader, nil)
	req, err := coder.Decode()
	if nil != err {
		t.Errorf("%s\n", err.Error())
	}

	if r, ok := req.(*MemRequest); !ok || r.Op != GET || r.Key != "a" {
		t.Errorf("parser fail\n")
	}
}

func Test_Delete(t *testing.T) {
	line := "delete abc\r\n"
	reader := strings.NewReader(line)
	coder := NewMemcachedServerCodec(reader, nil)
	req, err := coder.Decode()
	if nil != err {
		t.Errorf("%s\n", err.Error())
	}
	if r, ok := req.(*MemRequest); !ok || r.Op != DELETE || r.Key != "abc" {
		t.Errorf("parser fail\n")
	}
}

func Test_Incr(t *testing.T) {
	line := "incr a 20\r\n"
	reader := strings.NewReader(line)
	coder := NewMemcachedServerCodec(reader, nil)
	req, err := coder.Decode()
	if nil != err {
		t.Errorf("%s\n", err.Error())
	}
	if r, ok := req.(*MemRequest); !ok || r.Op != INCR || r.Key != "a" || r.Value != 20 {
		t.Errorf("parser fail\n")
	}
}

func Test_Multi(t *testing.T) {
	line := "incr a 20\r\nincr a 20\r\n"
	reader := strings.NewReader(line)
	coder := NewMemcachedServerCodec(reader, nil)
	for i := 0; i < 2; i++ {
		req, err := coder.Decode()
		if nil != err {
			t.Errorf("%s\n", err.Error())
		}
		if r, ok := req.(*MemRequest); !ok || r.Op != INCR || r.Key != "a" || r.Value != 20 {
			t.Errorf("parser fail\n")
		}
	}

}
