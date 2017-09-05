package sock

import (
	"net"
	"strconv"
	"sync"
	"testing"
	"time"
)

func Test_Accept_Timeout(t *testing.T) {
	ln, _ := net.Listen("tcp", "0.0.0.0:9990")
	listener, _ := ln.(*net.TCPListener)
	for i := 0; i < 3; i++ {
		tm := time.Now()
		tm = tm.Add(time.Millisecond * 1000)
		listener.SetDeadline(tm)
		_, err := listener.Accept()
		t.Logf("error = %s\n", err.Error())
	}
}

func Test_Server(t *testing.T) {
	address := "0.0.0.0:9991"
	server := NewServer(address)
	server.OnConnect(func(cn *Connection) {
		if !cn.IsClosed() {
			codec := LineCodecBuild(cn.GetTcpConn(), cn.GetTcpConn())
			cn.SetCodec(codec)
			for i := 0; i < 100; i++ {
				s := strconv.Itoa(i)
				cn.Send([]byte(s))
			}
		}
	})
	server.OnRead(func(cn *Connection, msg *Message) {
		buf := msg.Body.([]byte)
		cn.Send(buf)
	})
	err := server.Listen()
	if err != nil {
		t.Errorf("server listenr error = %s\n", err.Error())
		return
	}
	go func() {
		server.Start()
	}()

	client := NewClient(address /*, LineCodecBuild, option*/)
	var wg sync.WaitGroup
	wg.Add(100)
	client.OnConnect(func(cn *Connection) {
		if !cn.IsClosed() {
			codec := LineCodecBuild(cn.GetTcpConn(), cn.GetTcpConn())
			cn.SetCodec(codec)
		}
	})
	client.OnRead(func(cn *Connection, msg *Message) {
		wg.Done()
	})

	client.Connect()
	wg.Wait()
	server.Close()

}
