package memcached

import (
	"errors"
	"time"
	"github.com/golang/glog"
)

type MemcachedConfig struct {
	ConnectTimeoutMs     time.Duration
	ReadTimeoutMs        time.Duration
	ReConnectDelayMs     time.Duration
	OpTimeoutMs          time.Duration

	HeartBeatInterval time.Duration
	HeartBeatMaxRetries int

	ReadBufferSize       int
	WriteBufferSize      int
	PoolSize             int

	SendingQueueCapacity int
	SentQueueCapacity    int
}

type AddressInfo struct {
	Address string
	Weight  int
}

type MemcachedClient struct {
	addressList []*AddressInfo
	sessions    []*Session
	config      *MemcachedConfig
	connector   *Connector
}

func NewMemcachedClient(addressList []*AddressInfo) *MemcachedClient {
	c := new(MemcachedClient)
	c.addressList = addressList

	c.config = &MemcachedConfig{
		ConnectTimeoutMs: 5000 * time.Millisecond,
		ReadTimeoutMs:    5000 * time.Millisecond,
		ReConnectDelayMs: 5000 * time.Millisecond,
		OpTimeoutMs:      5000 * time.Millisecond,
		HeartBeatInterval: 6000 * time.Millisecond,
		HeartBeatMaxRetries: 3,
		ReadBufferSize:   16 * 1024,
		PoolSize:         2,
		SendingQueueCapacity : 1024,
		SentQueueCapacity : 1024,
	}
	c.connector = newConnector(addressList, c.config)
	return c
}

func (c *MemcachedClient) Start() {
	c.connector.start()
}

func (c *MemcachedClient) Close() {
	glog.Info("close %b", nil == c.connector)
	c.connector.close()
}

func (c *MemcachedClient) send(req *MemRequest) (*MemResponse, error) {
	key := req.Key
	s := c.connector.getSession(key)
	if nil == s || s.IsClosed() {
		return nil, errors.New("session is closed")
	}
	return s.send(req, c.config.OpTimeoutMs)
}

func (c *MemcachedClient) Get(key string) (string, error) {
	req := &MemRequest{Op: GET, Key: key}
	resp, err := c.send(req)
	if nil != err {
		return "", err
	}
	if resp.Result {
		return string(resp.Data), nil
	}
	return "", errors.New(resp.Err)
}
func (c *MemcachedClient) Set(key string, value string) (bool, error) {
	req := &MemRequest{Op: SET, Key: key, Data: []byte(value), Bytes: uint16(len(value)), Exptime: 0}
	resp, err := c.send(req)
	if nil != err {
		return false, err
	}
	if resp.Result {
		return true, nil
	}
	return false, errors.New(resp.Err)
}
func (c *MemcachedClient) Add(key string, value string) (bool, error) {
	req := &MemRequest{Op: ADD, Key: key, Data: []byte(value), Bytes: uint16(len(value)), Exptime: 0}
	resp, err := c.send(req)
	if nil != err {
		return false, err
	}
	if resp.Result {
		return true, nil
	}
	return false, errors.New(resp.Err)
}
func (c *MemcachedClient) Replace(key string, value string) (bool, error) {
	req := &MemRequest{Op: REPLACE, Key: key, Data: []byte(value), Bytes: uint16(len(value)), Exptime: 0}
	resp, err := c.send(req)
	if nil != err {
		return false, err
	}
	if resp.Result {
		return true, nil
	}
	return false, errors.New(resp.Err)
}
func (c *MemcachedClient) Delete(key string) (bool, error) {
	req := &MemRequest{Op: DELETE, Key: key}
	resp, err := c.send(req)
	if nil != err {
		return false, err
	}
	if resp.Result {
		return true, nil
	}
	return false, errors.New(resp.Err)
}

func (c *MemcachedClient) Incr(key string, value uint32) (uint32, error) {
	req := &MemRequest{Op: INCR, Key: key, Value: value}
	resp, err := c.send(req)
	if nil != err {
		return 0, err
	}
	if resp.Result {
		return resp.Value, nil
	}
	return 0, errors.New(resp.Err)
}
func (c *MemcachedClient) Decr(key string, value uint32) (uint32, error) {
	req := &MemRequest{Op: DECR, Key: key, Value: value}
	resp, err := c.send(req)
	if nil != err {
		return 0, err
	}
	if resp.Result {
		return resp.Value, nil
	}
	return 0, errors.New(resp.Err)
}
