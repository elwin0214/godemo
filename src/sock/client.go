package sock

import (
	"bufio"
	. "logger"
	"net"
)

type Client struct {
	address            string
	counter            uint32
	codecBuild         CodecBuild
	option             Option
	connectionCallBack ConnectionCallBack
	readCallBack       ReadCallBack
}

func NewClient(address string, codecBuild CodecBuild, option Option) *Client {
	client := new(Client)
	client.address = address
	client.counter = 0
	client.codecBuild = codecBuild
	client.option = option
	return client
}

func (c *Client) OnConnect(callback ConnectionCallBack) {
	c.connectionCallBack = callback
}

func (c *Client) OnRead(callback ReadCallBack) {
	c.readCallBack = callback
}

func (c *Client) Connect() error {
	cn, err := net.Dial("tcp", c.address)
	if err != nil {
		return err
	}
	tcpCon, _ := cn.(*net.TCPConn)
	tcpCon.SetNoDelay(c.option.NoDely)
	tcpCon.SetKeepAlive(c.option.KeepAlive)
	c.counter = c.counter + 1
	index := c.counter

	writer := bufio.NewWriterSize(tcpCon, c.option.WriteBufferSize)
	con := NewConnection(tcpCon, writer, index, c.codecBuild(tcpCon, writer))
	con.setConnectionCallBack(c.connectionCallBack)
	con.setReadCallBack(c.readCallBack)
	con.establish()
	go con.readLoop()
	go con.writeLoop()
	LOG.Info("[Connect] connect a new connection %s\n", con.name)
	return nil
}
