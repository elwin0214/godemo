package memcached

import (
	"bytes"
	. "memcached"
	"testing"
)

func Test_Client_Get(t *testing.T) {
	line := "get key\r\n"
	req := &MemRequest{Op: GET, Key: "key", Flags: 1, Exptime: 1}
	buffer := bytes.NewBuffer(make([]byte, 0, 1024))

	codec := NewMemcachedClientCodec(nil, buffer)
	err := codec.Encode(req)
	if nil != err {
		t.Errorf("error is %s\n", err.Error())
	}
	if line != string(buffer.Bytes()) {
		t.Errorf("%s\n", string(buffer.Bytes()))
	}
}

func Test_Client_Set(t *testing.T) {
	line := "set key 1 1 2\r\nab\r\n"
	req := &MemRequest{Op: SET, Key: "key", Flags: 1, Exptime: 1, Bytes: 2}
	req.Data = []byte("ab")
	buffer := bytes.NewBuffer(make([]byte, 0, 1024))

	codec := NewMemcachedClientCodec(nil, buffer)
	err := codec.Encode(req)
	if nil != err {
		t.Errorf("error is %s\n", err.Error())
	}
	if line != string(buffer.Bytes()) {
		t.Errorf("%s\n", string(buffer.Bytes()))
	}
}

func Test_Client_Decode(t *testing.T) {
	buffer := bytes.NewBuffer(make([]byte, 0, 102400))
	for i := 0; i < 1024; i++ {
		buffer.Write([]byte("STORED\r\n"))
	}
	codec := NewMemcachedClientCodec(buffer, nil)
	for i := 0; i < 1024; i++ {
		resp, _ := codec.Decode()
		if nil == resp {
			t.Errorf("can not decode a response\n")
		}
	}
	resp, _ := codec.Decode()
	if nil != resp {
		t.Errorf("can decode a response\n")
	}
}
