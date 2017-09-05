package memcached

import (
	"net/http"
	_ "net/http/pprof"
	"testing"
)

func Test_Client_SetGet(t *testing.T) {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()
	address := "127.0.0.1:9999"
	s := NewMemcachedServer(address, 32*1024)
	ch := make(chan bool, 1)
	go func() {
		s.Listen()
		ch <- true
		s.Start()
	}()
	<-ch
	c := NewMemcachedClient(address, 1, 5000)
	c.Start()

	key := "k1"
	value := "12"
	b, _ := c.Set(key, value)
	if !b {
		t.Errorf("set '%s' fail\n", key)
	}
	str, _ := c.Get(key)
	if str != value {
		t.Errorf("value is %s\n", str)
	}

	i, _ := c.Incr(key, 3)

	if 15 != i {
		t.Errorf("value is %d\n", i)
	}
	c.Close()
	s.Close()
}
