package sock

import (
	. "sock"
	"sync"
	"testing"
	"time"
)

func Test_Server_HeartBeat(t *testing.T) {
	option := Option{NoDely: true, KeepAlive: true, ReadBufferSize: 1024, WriteBufferSize: 1024}
	address := "127.0.0.1:9999"
	server := NewServer(address, LineCodecBuild, option)
	server.OnConnect(func(cn *Connection) {
		if cn.IsClosed() {
			t.Logf("connectin %s is closed.\n", cn.GetName())
		} else {
			cn.SetReadWriteChannelTimeout(500)
			cn.SetReadWriteChannelTimeoutCallBack(func(c *Connection) {
				c.Send([]byte("hello"))
			})
			t.Logf("connectin %s is connected.\n", cn.GetName())
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
	time.Sleep(time.Millisecond * 500)
	client := NewClient(address, LineCodecBuild, option)
	var wg sync.WaitGroup
	wg.Add(3)
	num := 3
	client.OnRead(func(cn *Connection, msg *Message) {
		buf := msg.Body.([]byte)
		text := string(buf)
		t.Logf("client receive '%s' from connection %s\n", text, cn.GetName())
		if num > 0 {
			wg.Done()
			num--
		}

	})
	client.Connect()
	wg.Wait()
	server.Close()
}
