package sock

import (
	. "logger"
	"net"
	"sync/atomic"
)

type Client struct {
	address            string
	counter            uint32
	codecBuild         CodecBuild
	connectionCallBack ConnectionCallBack
	readCallBack       ReadCallBack
}

func NewClient(address string, codecBuild CodecBuild) *Client {
	client := new(Client)
	client.address = address
	client.counter = 0
	client.codecBuild = codecBuild
	return client
}

func (c *Client) SetConnectionCallBack(callback ConnectionCallBack) {
	c.connectionCallBack = callback
}

func (c *Client) SetReadCallBack(callback ReadCallBack) {
	c.readCallBack = callback
}

func (c *Client) Connect() error {
	cn, err := net.Dial("tcp", c.address)
	if err != nil {
		return err
	}
	tcpCon, _ := cn.(*net.TCPConn)
	tcpCon.SetNoDelay(true)
	tcpCon.SetKeepAlive(true)
	index := atomic.AddUint32(&c.counter, 1)
	con := NewConnection(tcpCon, index, c.codecBuild(tcpCon, tcpCon))
	con.setConnectionCallBack(c.connectionCallBack)
	con.setReadCallBack(c.readCallBack)
	con.establish()
	go con.readLoop()
	go con.writeLoop()
	LOG.Info("[Connect] connect a new connection %s\n", con.name)
	return nil
}
