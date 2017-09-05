package memcached

import (
	"bufio"
	"context"
	. "github.com/elwin0214/gomemcached/sock"
	. "github.com/elwin0214/gomemcached/util"
	"github.com/golang/glog"
	"net"
)

type Command struct {
	req      *MemRequest
	respChan chan *MemResponse
}
type Session struct {
	conn         net.Conn
	writer       *bufio.Writer
	codec        Codec
	sendingQueue chan Command
	sentQueue    chan Command
	ctx          context.Context
	cancel       func()
	closeFlag    AtomicInt32
}

func NewSession(conn net.Conn, writeBufferSize int) *Session {
	s := new(Session)
	s.conn = conn
	s.writer = bufio.NewWriterSize(conn, writeBufferSize)
	s.codec = NewMemcachedClientCodec(conn, s.writer)
	s.sendingQueue = make(chan Command, 1024)
	s.sentQueue = make(chan Command, 1024)
	s.closeFlag = 0
	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancel = cancel
	s.start()
	return s
}

func (s *Session) close() {
	glog.Infof("[close] goto close session\n")
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
			glog.Errorf("[readLoop] error = %s\n", err.Error())
			s.close()
			return
		}
		select {
		case cmd := <-s.sentQueue:
			mresp, _ := resp.(*MemResponse)
			cmd.respChan <- mresp
		default: //not exit sent cmd
			glog.Errorf("[readLoop] not exist sent command for %v\n", resp)
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
		}
	}
}
func (s *Session) send(req *MemRequest) *MemResponse {
	glog.Infof("[send] op = %d key = %s \n", req.Op, req.Key)
	msg := Command{req: req, respChan: make(chan *MemResponse, 1)}
	//dont block in sentQueue when closing
	select {
	case s.sendingQueue <- msg:
		return <-msg.respChan
	case <-s.ctx.Done():
		return nil
	}

}
