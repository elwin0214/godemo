package memcached

import (
	"net/http"
	_ "net/http/pprof"
	"testing"
	"time"
	"github.com/golang/glog"
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
	c := NewMemcachedClient([]*AddressInfo{&AddressInfo{Address:"127.0.0.1:9999",Weight:1}})
	c.Start()
	time.Sleep(2000*time.Millisecond)
	key := "k1"
	value := "12"
	b, e := c.Set(key, value)
	if !b {
		t.Errorf("set '%s' fail error is '%s'\n", key, e.Error())
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
	glog.Info("close")
	time.Sleep(1*time.Second)
	s.Close()
}
