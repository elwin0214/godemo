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
	c.client = NewClient(address, codecBuild)
	c.interval = interval
	c.client.SetConnectionCallBack(func(cn *Connection) {
		if !cn.IsClosed() {
			c.onConnection(cn)
		}
	})
	return c
}

func (hbc *HeartBeatClient) Connect() error {
	return hbc.client.Connect()
}

func (hbc *HeartBeatClient) onConnection(cn *Connection) {
	cn.SetReadWriteChannelTimeout(hbc.interval)
	cn.SetReadWriteChannelTimeoutCallBack(func(c *Connection) {
		c.Send([]byte("HELLO"))
	})
}
