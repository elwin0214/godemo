package sock

import (
	"net"
	"syscall"
	//"reflect"
	"testing"
	//"unsafe"
)

func Test_Bind_AfterClose(t *testing.T) {
	address := "127.0.0.1:8887"
	lnCh := make(chan bool, 1)
	closeServerCh := make(chan bool, 1)
	closeClientCh := make(chan bool, 1)

	go func() {
		ln, err := net.Listen("tcp", address)
		if err != nil {
			t.Errorf("listen error = %s\n", err.Error())
		}
		listener, lerr := ln.(*net.TCPListener)
		if !lerr {
			t.Errorf("cast fail\n")
		}
		file, ferr := listener.File()
		if ferr != nil {
			t.Errorf("listen error = %s\n", ferr.Error())
		}
		fd := int(file.Fd())
		t.Logf("fd = %d\n", fd)
		//fd := (*int)(unsafe.Pointer(fdptr))
		//t.Logf("fd = %d\n", fd)
		//t.Logf("%v\n", reflect.TypeOf(fd))
		op, _ := syscall.GetsockoptInt((fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR)
		t.Logf("fd = %d op = %d\n", fd, op)

		lnCh <- true
		cn, e := ln.Accept()
		if e != nil {
			t.Errorf("accept error = %s\n", err.Error())
		}
		t.Logf("accept a connection\n")
		cn.Close()
		ln.Close()
		closeServerCh <- true
	}()

	go func() {
		<-lnCh
		_, err := net.Dial("tcp", address)
		if err != nil {
			t.Errorf("%s\n", err.Error())
			return
		}
		<-closeClientCh
	}()
	<-closeServerCh
	closeClientCh <- true
	ln, err := net.Listen("tcp", "127.0.0.1:8888")
	if err != nil {
		t.Errorf("error = %s\n", err.Error())
	}
	ln.Close()
}
