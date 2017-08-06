package memcached

import (
	"errors"
	. "logger"
	"math/rand"
	"net"
	"time"
)

type MemcachedClient struct {
	address          string
	sessions         []*Session
	connectTimeoutMs time.Duration
	readTimeoutMs    time.Duration
	sessionSize      int
}

func NewMemcachedClient(address string, sessionSize int, connectTimeoutMs time.Duration) *MemcachedClient {
	c := new(MemcachedClient)
	c.address = address
	c.sessionSize = sessionSize
	c.connectTimeoutMs = connectTimeoutMs
	return c
}

func (c *MemcachedClient) Start() {
	c.sessions = make([]*Session, 0, c.sessionSize)
	for i := 0; i < c.sessionSize; i++ {
		con, _ := c.connect()
		c.sessions = append(c.sessions, NewSession(con, 16*1024))
	}
}

func (c *MemcachedClient) Close() {
	for i := 0; i < len(c.sessions); i++ {
		LOG.Debug("[Close] session = %d\n", i)
		c.sessions[i].close()
	}

}

func (c *MemcachedClient) connect() (net.Conn, error) {
	cn, err := net.DialTimeout("tcp", c.address, time.Millisecond*c.connectTimeoutMs)
	if err != nil {
		return cn, err
	}
	tcpCon, _ := cn.(*net.TCPConn)
	tcpCon.SetKeepAlive(true)
	tcpCon.SetNoDelay(true)
	return cn, nil
}

func (c *MemcachedClient) send(req *MemRequest) *MemResponse {
	s := c.sessions[rand.Intn(c.sessionSize)]
	return s.send(req)
}

func (c *MemcachedClient) Get(key string) (string, error) {
	req := &MemRequest{Op: GET, Key: key}
	resp := c.send(req)
	if nil == resp {
		return "", errors.New("session is closed")
	}
	if resp.Result {
		return string(resp.Data), nil
	}
	return "", errors.New(resp.Err)
}
func (c *MemcachedClient) Set(key string, value string) (bool, error) {
	req := &MemRequest{Op: SET, Key: key, Data: []byte(value), Bytes: uint16(len(value)), Exptime: 0}
	resp := c.send(req)
	if nil == resp {
		return false, errors.New("session is closed")
	}
	if resp.Result {
		return true, nil
	}
	return false, errors.New(resp.Err)
}

func (c *MemcachedClient) Delete(key string) (bool, error) {
	req := &MemRequest{Op: DELETE, Key: key}
	resp := c.send(req)
	if nil == resp {
		return false, errors.New("session is closed")
	}
	if resp.Result {
		return true, nil
	}
	return false, errors.New(resp.Err)
}

func (c *MemcachedClient) Incr(key string, value uint32) (uint32, error) {
	req := &MemRequest{Op: INCR, Key: key, Value: value}
	resp := c.send(req)
	if nil == resp {
		return 0, errors.New("session is closed")
	}
	if resp.Result {
		return resp.Value, nil
	}
	return 0, errors.New(resp.Err)
}
func (c *MemcachedClient) Decr(key string, value uint32) (uint32, error) {
	req := &MemRequest{Op: DECR, Key: key, Value: value}
	resp := c.send(req)
	if nil == resp {
		return 0, errors.New("session is closed")
	}
	if resp.Result {
		return resp.Value, nil
	}
	return 0, errors.New(resp.Err)
}
