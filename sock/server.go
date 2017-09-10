package sock

import (
	. "github.com/elwin0214/gomemcached/util"
	"github.com/golang/glog"
	"net"
	"time"
)

type Server struct {
	address         string
	listener        *net.TCPListener
	counter         uint32
	closed          AtomicBool
	connectCallBack ConnectionCallBack
	closeCallBack   ConnectionCallBack
	readCallBack    ReadCallBack
}

func NewServer(address string) *Server {
	s := &Server{address: address}
	s.counter = 0
	s.closed = NewAtomicBool(true)
	return s
}

func (s *Server) OnConnect(callback ConnectionCallBack) {
	s.connectCallBack = callback
}

func (s *Server) onClose(callback ConnectionCallBack) {
	s.closeCallBack = callback
}

func (s *Server) OnRead(callback ReadCallBack) {
	s.readCallBack = callback
}

func (s *Server) Listen() error {

	ln, err := net.Listen("tcp", s.address)
	if err != nil {
		glog.Errorf("[Start] error = %s\n", err.Error())
		return err
	}

	glog.Infof("[Listen] address = %s\n", s.address)
	s.listener, _ = ln.(*net.TCPListener)
	s.closed.Set(false)
	return nil
}

func (s *Server) Start() {

	for {
		if s.closed.Get() {
			return
		}
		t := time.Now()
		t = t.Add(time.Millisecond * 5000)
		s.listener.SetDeadline(t)
		cn, acceptErr := s.listener.Accept()
		if acceptErr != nil {
			//glog.Infof("[Start] accept error = %s\n", acceptErr.Error())
			continue
		}
		tcpCon, _ := cn.(*net.TCPConn)
		tcpCon.SetNoDelay(true)
		tcpCon.SetKeepAlive(true)
		s.counter = s.counter + 1
		index := s.counter
		con := NewConnection(tcpCon, index)
		con.setConnectCallBack(s.connectCallBack)
		con.setCloseCallBack(s.closeCallBack)
		con.setReadCallBack(s.readCallBack)
		con.establish()

		go con.readLoop()
		go con.writeLoop()
		glog.Infof("[Start] accept a new connection name = %s, id = %d\n", con.name, con.id)
	}
}

func (s *Server) Close() {
	if s.closed.Cas(false, true) {
		glog.Infof("[Close] server closed")
		s.listener.Close()
	}
}
