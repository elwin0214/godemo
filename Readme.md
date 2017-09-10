It is a project to learn golang.

### Design
The memcached client/server only support partial text protocol.  

In server side, for each connection there will
be two goroutes, one for reading and another for writing. The ```Codec``` is used to encode the message and decode the buffer. Besides another goroute is used to receive   ```*MemRequest``` and send ```*MemResponse```
after handled the request.   

In client side, the ```Session``` represents a TCP conection. For each Session, there are three goroutes, two for reading and writing, another for heartbeat. The caller will not write the ```*MemRequest``` to the socket directly, It will send the ```*MemRequest```  to the channel in ```Session```, then the writing goroute will collect requests as many as possiable and write to the socket.

It provide a simple TCP framework.

### Example

* echo server
```golang
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
```

* echo client
```golang
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
```

* [Memcached Server(pprof)](https://github.com/elwin0214/gomemcached/blob/master/main/mem_server.go)
* [Memcached Client(pprof)](https://github.com/elwin0214/gomemcached/blob/master/main/mem_client.go)

### Build
```shell
go get github.com/golang/glog
mkdir <GOPATH>/src/elwin0214/
git clone xxx
cd gomemcached
make test|bench|build|clean
```

### Memcached protocol
```
https://github.com/memcached/memcached/blob/master/doc/protocol.txt
```



