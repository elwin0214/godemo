package memcached

import (
	"bufio"
	"context"
	. "github.com/elwin0214/gomemcached/sock"
	. "github.com/elwin0214/gomemcached/util"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"net"
	"sync"
	"time"
)

type Command struct {
	req      *MemRequest
	fun      func()
	respChan chan *MemResponse
}
type Session struct {
	id          string       // local host:port
	addressInfo *AddressInfo //remote address

	conn   net.Conn
	writer *bufio.Writer
	codec  Codec

	sendingQueue chan Command
	sentQueue    chan Command

	lastOpTime  time.Time // for heartbeat
	retries     int
	heartBeatMu sync.Mutex

	wakeupTimer *time.Timer
	wakeupMu    sync.Mutex

	config *MemcachedConfig

	ctx    context.Context
	cancel func()
	closed AtomicBool

	connector *Connector
}

func NewSession(conn net.Conn, sendingQueueCapcity, sentQueueCapcity, writeBufferSize int) *Session {
	s := new(Session)
	s.conn = conn
	s.id = conn.LocalAddr().String()
	s.writer = bufio.NewWriterSize(conn, writeBufferSize)
	s.codec = NewMemcachedClientCodec(conn, s.writer)
	s.sendingQueue = make(chan Command, sendingQueueCapcity)
	s.sentQueue = make(chan Command, sentQueueCapcity)
	s.closed = NewAtomicBool(false)
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.lastOpTime = time.Now()
	s.wakeupTimer = time.NewTimer(10 * time.Second)
	s.start()
	return s
}

func NewClosedSession() *Session {
	s := new(Session)
	s.closed = NewAtomicBool(true)
	return s
}

func (s *Session) setConnector(connector *Connector) {
	s.connector = connector
}
func (s *Session) setConfig(cfg *MemcachedConfig) {
	s.config = cfg
}
func (s *Session) remoteAddress() string {
	return s.addressInfo.Address
}
func (s *Session) setAddress(addressInfo *AddressInfo) {
	s.addressInfo = addressInfo
}

func (s *Session) IsClosed() bool {
	return s.closed.Get()
}

func (s *Session) close() {
	glog.InfoDepth(1, "[close] goto close session\n")
	if s.closed.Cas(false, true) {
		s.conn.Close()
		s.cancel() //will wakeup the goroute which invoke the set/get/....
		s.connector.onClose(s)
	}
}

func (s *Session) start() {
	go s.readLoop()
	go s.writeLoop(s.ctx)
	go s.onHeartBeat(s.ctx)
}

func (s *Session) wakeup() {
	s.wakeupMu.Lock()
	defer s.wakeupMu.Unlock()
	if !s.wakeupTimer.Stop() {
		select {
		case <-s.wakeupTimer.C:
		default:
		}
	}
	s.wakeupTimer.Reset(0 * time.Second)
}
func (s *Session) readLoop() {
	for {
		resp, err := s.codec.Decode()
		if nil != err {
			glog.Errorf("[readLoop] error = %s\n", err.Error())
			s.close()
			return
		}
		s.resetOpTime(true)
		select {
		case cmd := <-s.sentQueue:
			mresp, _ := resp.(*MemResponse)
			if nil != cmd.fun {
				cmd.fun()
			}
			cmd.respChan <- mresp
		default:
			//not exit sent cmd
			glog.Errorf("[readLoop] not exist sent command for %v\n", resp)
			s.close()
		}
	}
}

func (s *Session) writeLoop(ctx context.Context) {
	for {
		select {
		case <-s.wakeupTimer.C:
			break
		case <-ctx.Done():
			return
		}
	L1:
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
				glog.Infof("[writeLoop] op = %d key = %s \n", cmd.req.Op, cmd.req.Key)
			}

			if got {
				//dont block in sentQueue when closing
				select {
				case s.sentQueue <- cmd:
					err := s.codec.Encode(cmd.req)
					if nil != err {
						glog.Errorf("[writeLoop] error = %s\n", err.Error())
						s.close()
						return
					}

				case <-ctx.Done():
					s.close()
					return
				}
			} else {
				//LOG.Debug("[writeLoop] buffer = %d\n", s.writer.Buffered())
				err := s.writer.Flush()
				if nil != err {
					glog.Errorf("[writeLoop] error = %s\n", err.Error())
					s.close()
					return
				}
				break L1
			}
		}
	}
}

func (s *Session) send(req *MemRequest, timeout time.Duration) (*MemResponse, error) {
	glog.Infof("[send] op = %d key = %s", req.Op, req.Key)
	msg := Command{req: req, respChan: make(chan *MemResponse, 1)} // avoid blocking
	//dont block in sentQueue when closing
	select {
	case s.sendingQueue <- msg:
		s.wakeup()
		return <-msg.respChan, nil
	case <-time.NewTimer(timeout).C:
		return nil, errors.New("session timeout")
	case <-s.ctx.Done():
		return nil, errors.New("session will be closed")
	}
}

func (s *Session) resetOpTime(succ bool) bool {
	s.heartBeatMu.Lock()
	defer s.heartBeatMu.Unlock()
	s.lastOpTime = time.Now()
	if succ {
		s.retries = 0
		return false
	}
	if s.retries < s.config.HeartBeatMaxRetries {
		s.retries = s.retries + 1
	}
	return s.retries >= s.config.HeartBeatMaxRetries

}

func (s *Session) onHeartBeat(ctx context.Context) {
	for !s.closed.Get() {
		now := time.Now()
		var delay time.Duration
		s.heartBeatMu.Lock()
		delay = s.lastOpTime.Add(s.config.HeartBeatInterval).Sub(now)
		s.heartBeatMu.Unlock()

		if delay <= 0 {
			s.onHeartBeat0()
		} else {
			select {
			case <-time.NewTimer(delay).C:
				continue
			case <-ctx.Done():
				return
			}
		}
	}
}

func (s *Session) onHeartBeat0() {
	req := &MemRequest{Op: VER}
	_, err := s.send(req, s.config.OpTimeoutMs)
	toclose := s.resetOpTime(nil == err)
	if toclose {
		s.close()
	}
}
