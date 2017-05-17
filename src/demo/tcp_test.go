package demo

import (
	"net"
	"testing"
	"time"
)

func Test_TcpReadTimeout(t *testing.T) {
	address := "127.0.0.1:9999"
	listener, err := net.Listen("tcp", address)
	if err != nil {
		t.Errorf("%s\n", err.Error())
		return
	}
	ch := make(chan bool, 1)
	go func() {
		net.Dial("tcp", address)
		<-ch
	}()

	cn, ae := listener.Accept()
	if ae != nil {
		t.Errorf("%s\n", ae.Error())
		return
	}

	cn.SetReadDeadline(time.Now().Add(time.Millisecond * 2000))

	buf := make([]byte, 1024)
	_, re := cn.Read(buf)
	if nil != re {
		t.Logf("%s\n", re.Error())
		cn.Close()
	}
	ch <- true
}
