package sock

import (
	. "sock"
	"strconv"
	"sync"
	"testing"
)

func Test_Server(t *testing.T) {
	option := Option{NoDely: true, KeepAlive: true, ReadBufferSize: 1024, WriteBufferSize: 1024}
	address := "127.0.0.1:9999"
	server := NewServer(address, LineCodecBuild, option)
	server.OnConnect(func(cn *Connection) {
		if !cn.IsClosed() {
			for i := 0; i < 100; i++ {
				s := strconv.Itoa(i)
				cn.Send([]byte(s))
			}
		}
	})
	server.OnRead(func(cn *Connection, msg *Message) {
		buf := msg.Body.([]byte)
		//text := string(buf)
		//t.Logf("server receive '%d' from connection %s ", (text), cn.GetName())
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

	client := NewClient(address, LineCodecBuild, option)
	var wg sync.WaitGroup
	wg.Add(100)
	client.OnRead(func(cn *Connection, msg *Message) {
		//buf := msg.Body.([]byte)
		//text := string(buf)
		//i, _ := strconv.Atoi(text)
		//t.Logf("client receive '%d' from connection %d ", i, cn.GetName())
		wg.Done()
	})

	client.Connect()
	wg.Wait()
	server.Close()

}
