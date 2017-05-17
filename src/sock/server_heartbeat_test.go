package sock

import (
	. "sock"
	"sync"
	"testing"
	"time"
)

func Test_Server_HeartBeat(t *testing.T) {
	address := "127.0.0.1:9999"
	server := NewServer(address, LineCodecBuild)
	server.SetConnectionCallBack(func(cn *Connection) {
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
	time.Sleep(time.Millisecond * 2000)
	client := NewClient(address, LineCodecBuild)
	var wg sync.WaitGroup
	wg.Add(10)
	num := 10
	client.SetReadCallBack(func(cn *Connection, msg *Message) {
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
