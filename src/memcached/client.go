package memcached

import (
	"bufio"
	"context"
	"errors"
	. "logger"
	"math/rand"
	"net"
	. "sock"
	"time"
	"util"
)

type Command struct {
	req      *MemRequest
	respChan chan *MemResponse
}
type Session struct {
	conn   net.Conn
	writer *bufio.Writer
	//		writer := bufio.NewWriterSize(tcpCon, s.option.WriteBufferSize)

	codec        Codec
	sendingQueue chan Command
	sentQueue    chan Command
	ctx          context.Context
	cancel       func()
	closeFlag    *util.AtomicInt
}

func newSession(conn net.Conn, writeBufferSize int) *Session {
	s := new(Session)
	s.conn = conn
	s.writer = bufio.NewWriterSize(conn, writeBufferSize)
	s.codec = NewMemcachedClientCodec(conn, s.writer)
	s.sendingQueue = make(chan Command, 1024)
	s.sentQueue = make(chan Command, 1024)
	s.closeFlag = util.NewAtomicInt(0)
	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancel = cancel
	s.start()
	return s
}

func (s *Session) close() {
	LOG.Debug("[close] goto close session\n")
	if s.closeFlag.Cas(0, 1) {
		s.conn.Close()
		s.cancel()
		//dont block the goroute which invoke the set/get/....
	L1:
		for {

			select {
			case cmd := <-s.sendingQueue:
				cmd.respChan <- nil // todo block ??
			default:
				break L1
			}
		}
	L2:
		for {
			select {
			case cmd := <-s.sentQueue:
				cmd.respChan <- nil // todo block ??
			default:
				break L2
			}
		}

	}
}

func (s *Session) start() {
	go s.readLoop()
	go s.writeLoop(s.ctx)
}

func (s *Session) readLoop() {
	for {
		resp, err := s.codec.Decode()
		if nil != err {
			LOG.Error("[writeLoop] error = %s\n", err.Error())
			s.close()
			return
		}
		select {
		case cmd := <-s.sentQueue:
			mresp, _ := resp.(*MemResponse)
			cmd.respChan <- mresp
		default: //not exit sent cmd
			LOG.Error("[writeLoop] not exist sent command for %v\n", resp)
			s.close()
		}
	}
}

func (s *Session) writeLoop(ctx context.Context) {
	for {
		var cmd Command
		got := true
		//transfer the cmd from sending queue to the sent queue at first.
		//then write the request in cmd to the socket
		//otherwise the readLoop goroute can not find the cmd in the sent queue when got the response
		select {
		//dont block in sentQueue when closing
		case cmd = <-s.sendingQueue:

		case <-ctx.Done():
			return
		default:
			got = false

		}
		if got {
			LOG.Debug("[writeLoop] op = %d key = %s \n", cmd.req.Op, cmd.req.Key)
		}

		if got {
			//dont block in sentQueue when closing
			select {
			case s.sentQueue <- cmd:
				err := s.codec.Encode(cmd.req)
				if nil != err {
					LOG.Error("[writeLoop] error = %s\n", err.Error())
					s.close()
					return
				}

			case <-ctx.Done():
				s.close()
				return
			}
		} else {
			LOG.Debug("[writeLoop] buffer = %d\n", s.writer.Buffered())
			err := s.writer.Flush()
			if nil != err {
				LOG.Error("[writeLoop] error = %s\n", err.Error())
				s.close()
				return
			}
		}
	}
}
func (s *Session) send(req *MemRequest) *MemResponse {
	LOG.Debug("[send] op = %d key = %s \n", req.Op, req.Key)
	msg := Command{req: req, respChan: make(chan *MemResponse, 1)}
	//dont block in sentQueue when closing
	select {
	case s.sendingQueue <- msg:
		return <-msg.respChan
	case <-s.ctx.Done():
		return nil
	}

}

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
		c.sessions = append(c.sessions, newSession(con, 16*1024))
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
