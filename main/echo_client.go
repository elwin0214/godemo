package main

import (
	. "github.com/elwin0214/gomemcached/sock"
	"github.com/golang/glog"
	"sync"
	"flag"
)

func main() {
	flag.Parse()
	address := "0.0.0.0:9991"
	client := NewClient(address)
	var wg sync.WaitGroup
	wg.Add(1)
	client.OnConnect(func(cn *Connection) {
		codec := LineCodecBuild(cn.GetTcpConn(), cn.GetTcpConn())
		cn.SetCodec(codec)
		cn.Send([]byte("hello!"))
	})
	client.OnRead(func(cn *Connection, msg *Message) {
		body ,_:= msg.Body.([]byte)
		glog.Errorf("receive %s", string(body))
		cn.Close()
		wg.Done()
	})
	client.Connect()
	wg.Wait()
}
