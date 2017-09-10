package main

import (
	. "github.com/elwin0214/gomemcached/sig"
	. "github.com/elwin0214/gomemcached/sock"
	"github.com/golang/glog"
	"flag"
)

func main() {
	flag.Parse()
	address := "0.0.0.0:9991"
	server := NewServer(address)
	server.OnConnect(func(cn *Connection) {
		codec := LineCodecBuild(cn.GetTcpConn(), cn.GetTcpConn())
		cn.SetCodec(codec)
	})
	server.OnRead(func(cn *Connection, msg *Message) {
		buf := msg.Body.([]byte)
		glog.Errorf("receive '%s' from %s\n", string(buf), cn.GetName())
		cn.Send(buf)
	})
	err := server.Listen()
	if err != nil {
		glog.Errorf("server listenr error = %s\n", err.Error())
		return
	}
	RegisterStopSignal(func() {
		glog.Infof("close")
		server.Close()
	})
	server.Start()

}
