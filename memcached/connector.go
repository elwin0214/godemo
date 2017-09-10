package memcached

import (
	"context"
	"github.com/golang/glog"
	. "github.com/elwin0214/gomemcached/util"
	"net"
	"sync"
	"time"
)

type ReConnectTask struct {
	address string
}

type Connector struct {
	addressList []*AddressInfo
	sessionMap  map[string]*List
	config      *MemcachedConfig
	ctx         context.Context
	cancel      func()
	closed      AtomicBool
	mu          sync.Mutex
	smu         sync.Mutex
	locator     SessionLocator
}

func newConnector(addressList []*AddressInfo, config *MemcachedConfig) *Connector {
	c := new(Connector)
	c.addressList = addressList
	c.sessionMap = make(map[string]*List)
	c.config = config
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.closed = NewAtomicBool(true)
	c.locator = newArraySessionLocator(addressList)
	return c
}



func (c *Connector) start() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed.Set(false)
	for _, address := range c.addressList {
		for i := 0; i < c.config.PoolSize; i++ {
			c.connect(address)
		}
	}
}

func (c *Connector) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed.Set(true)
	c.cancel()
	c.closeSessions()
}

func (c *Connector) closeSessions() {
	c.smu.Lock()
	defer c.smu.Unlock()
	for _, sessionList := range c.sessionMap {
		for _, session := range *sessionList {
			s, _ := session.(*Session)
			s.close()
		}
	}
}

func (c *Connector) connect(addressInfo *AddressInfo) {
	cn, err := net.DialTimeout("tcp", addressInfo.Address, c.config.ConnectTimeoutMs)
	if err != nil {
		glog.Errorf("[connect] address = %s err = %s\n", addressInfo.Address, err.Error())
		c.connectLater(addressInfo)
		return
	}
	tcpCon, _ := cn.(*net.TCPConn)
	tcpCon.SetKeepAlive(true)
	tcpCon.SetNoDelay(true)
	glog.Infof("[connect] remote = %s local = %s\n", addressInfo.Address, tcpCon.LocalAddr().String())
	s := NewSession(tcpCon, c.config.SendingQueueCapacity, c.config.SentQueueCapacity, c.config.WriteBufferSize)
	s.setAddress(addressInfo)
	s.setConnector(c)
	s.setConfig(c.config)
	c.addSession(s)
}

func (c *Connector) getSession(key string) *Session {
	return c.locator.getSessionByKey(key)
}

func (c *Connector) removeSession(s *Session) {
	c.smu.Lock()
	defer c.smu.Unlock()
	//todo
	sessionList := c.sessionMap[s.addressInfo.Address]
	sessionList.Remove(s, func(e1, e2 interface{}) bool {
		s1, _ := e1.(*Session)
		s2, _ := e2.(*Session)
		return s1.id == s2.id
	});
	c.locator.updateSessions(c.sessionMap)
}

func (c *Connector) onClose(s *Session) {
	if !c.closed.Get() {
		c.removeSession(s)
		c.connectLater(s.addressInfo)
	}
}
func (c *Connector) addSession(s *Session) {
	c.smu.Lock()
	defer c.smu.Unlock()
	sessionList := c.sessionMap[s.addressInfo.Address]
	if nil == sessionList {
		tmpList := NewList(16);
		sessionList = &tmpList
		c.sessionMap[s.addressInfo.Address] = sessionList

	}
	sessionList.Append(s);
	c.locator.updateSessions(c.sessionMap) // todo
}

func (c *Connector) connectLater(addressInfo *AddressInfo) {
	go func(addressInfo *AddressInfo) {
		select {
		case <-time.NewTimer(c.config.ReConnectDelayMs).C:
			c.connect(addressInfo)
		case <-c.ctx.Done():
			return
		}

	}(addressInfo)
}
