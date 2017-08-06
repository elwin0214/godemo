package sock

import (
	"net"
	. "sock"
	"testing"
	"time"
)

func Test_Server_Timeout(t *testing.T) {
	address := "0.0.0.0:9999"
	server := NewServer(address)
	server.OnConnect(func(cn *Connection) {
		if cn.IsClosed() {
			t.Logf("connectin %s is closed.", cn.GetName())
		} else {
			codec := LineCodecBuild(cn.GetTcpConn(), cn.GetTcpConn())
			cn.SetCodec(codec)
			cn.SetReadTimeout(500)
			t.Logf("connectin %s is connected.", cn.GetName())
		}
	})
	err := server.Listen()
	if err != nil {
		t.Errorf("server listenr error = %s\n", err.Error())
		return
	}
	go func() {
		server.Start()
	}()
	_, err = net.Dial("tcp", address)
	if err != nil {
		t.Errorf("%s\n", err.Error())
		return
	}
	time.Sleep(time.Millisecond * 2000)
	server.Close()

}
