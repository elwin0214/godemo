package sock

import (
	"net"
	. "sock"
	"testing"
	"time"
)

func Test_Server_Timeout(t *testing.T) {

	address := "127.0.0.1:9999"
	server := NewServer(address, LineCodecBuild)
	server.SetConnectionCallBack(func(cn *Connection) {
		if cn.IsClosed() {
			t.Logf("connectin %s is closed.", cn.GetName())
		} else {
			cn.SetReadTimeout(1000)
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
	time.Sleep(time.Millisecond * 6000)
	server.Close()

}
