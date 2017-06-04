package heartbeat

import (
	. "sock"
	"time"
)

type HeartBeatClient struct {
	client   *Client
	interval time.Duration
}

func NewHeartBeatClient(address string, codecBuild CodecBuild, interval time.Duration) *HeartBeatClient {
	c := new(HeartBeatClient)
	option := Option{NoDely: true, KeepAlive: true, ReadBufferSize: 1024, WriteBufferSize: 1024}
	c.client = NewClient(address, codecBuild, option)
	c.interval = interval
	c.client.OnConnect(func(cn *Connection) {
		if !cn.IsClosed() {
			c.onConnect(cn)
		}
	})
	return c
}

func (hbc *HeartBeatClient) Connect() error {
	return hbc.client.Connect()
}

func (hbc *HeartBeatClient) onConnect(cn *Connection) {
	cn.SetReadWriteChannelTimeout(hbc.interval)
	cn.SetReadWriteChannelTimeoutCallBack(func(c *Connection) {
		c.Send([]byte("HELLO"))
	})
}
