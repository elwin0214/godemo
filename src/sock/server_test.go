package sock

import (
	. "sock"
	"strconv"
	"sync"
	"testing"
)

func Test_Server(t *testing.T) {

	address := "127.0.0.1:9999"
	server := NewServer(address, LineCodecBuild)
	server.SetConnectionCallBack(func(cn *Connection) {
		if cn.IsClosed() {
			t.Logf("connectin %s is closed.", cn.GetName())
		} else {
			t.Logf("connectin %s is connected.", cn.GetName())
		}

	})
	server.SetReadCallBack(func(cn *Connection, msg *Message) {
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

	client := NewClient(address, LineCodecBuild)
	var wg sync.WaitGroup
	wg.Add(100)
	client.SetReadCallBack(func(cn *Connection, msg *Message) {
		//buf := msg.Body.([]byte)
		//text := string(buf)
		//i, _ := strconv.Atoi(text)
		//t.Logf("client receive '%d' from connection %d ", i, cn.GetName())
		wg.Done()
	})

	server.SetConnectionCallBack(func(cn *Connection) {
		if !cn.IsClosed() {
			for i := 0; i < 100; i++ {
				s := strconv.Itoa(i)
				cn.Send([]byte(s))
			}
		}
	})
	client.Connect()
	wg.Wait()
	server.Close()

}
