It is a project to learn golang.

### Design
The memcached client/server only support partial text protocol.  

In server side, for each connection there will
be two goroutes, one for reading and another for writing. The ```Codec``` is used to encode the message and decode the buffer. Besides another goroute is used to receive   ```*MemRequest``` and send ```*MemResponse```
after handled the request.   

In client side, the ```Session``` represents a TCP conection. For each Session, there are three goroutes, two for reading and writing, another for heartbeat. The caller will not write the ```*MemRequest``` to the socket directly, It will send the ```*MemRequest```  to the channel in ```Session```, then the writing goroute will collect requests as many as possiable and write to the socket.

It provide a simple TCP framework.

### Example

echo server
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

echo client
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



