package sock

import (
	. "logger"
	"net"
	"time"
	. "util"
)

type Server struct {
	address  string
	listener *net.TCPListener
	counter  uint32
	//codecBuild         CodecBuild
	closeFlag AtomicInt32
	//option             Option
	connectionCallBack ConnectionCallBack
	readCallBack       ReadCallBack
}

func NewServer(address string /*, codecBuild CodecBuild, option Option*/) *Server {
	s := &Server{address: address}
	s.counter = 0
	s.closeFlag = 0
	//s.codecBuild = codecBuild
	//s.option = option
	return s
}

func (s *Server) OnConnect(callback ConnectionCallBack) {
	s.connectionCallBack = callback
}

func (s *Server) OnRead(callback ReadCallBack) {
	s.readCallBack = callback
}

func (s *Server) Listen() error {
	ln, err := net.Listen("tcp", s.address)
	if err != nil {
		LOG.Error("[Start] error = %s\n", err.Error())
		return err
	}

	LOG.Info("[Listen] address = %s\n", s.address)
	s.listener, _ = ln.(*net.TCPListener)
	return nil
}

func (s *Server) Start() {
	for {
		if s.closeFlag.Get() == 1 {
			return
		}
		t := time.Now()
		t = t.Add(time.Millisecond * 5000)
		s.listener.SetDeadline(t)
		cn, acceptErr := s.listener.Accept()
		if acceptErr != nil {
			LOG.Info("[Start] accept error = %s\n", acceptErr.Error())
			continue
		}
		tcpCon, _ := cn.(*net.TCPConn)
		tcpCon.SetNoDelay(true)
		tcpCon.SetKeepAlive(true)
		s.counter = s.counter + 1
		index := s.counter
		con := NewConnection(tcpCon, index)
		con.setConnectionCallBack(s.connectionCallBack)
		con.setReadCallBack(s.readCallBack)
		con.establish()

		go con.readLoop()
		go con.writeLoop()
		LOG.Info("[Start] accept a new connection name = %s, id = %d\n", con.name, con.id)
	}
}

func (s *Server) Close() {
	if s.closeFlag.Cas(0, 1) {
		LOG.Info("[Close] server closed")
		s.listener.Close()
	}
}
