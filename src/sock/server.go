package sock

import (
	. "logger"
	"net"
	"sync/atomic"
)

type Server struct {
	address    string
	listener   *net.TCPListener
	counter    uint32
	codecBuild CodecBuild
	closeFlag  int32

	connectionCallBack ConnectionCallBack
	readCallBack       ReadCallBack
}

func NewServer(address string, codecBuild CodecBuild) *Server {
	server := &Server{address: address}
	server.counter = 0
	server.closeFlag = 0
	server.codecBuild = codecBuild
	return server
}

func (s *Server) SetConnectionCallBack(callback ConnectionCallBack) {
	s.connectionCallBack = callback
}

func (s *Server) SetReadCallBack(callback ReadCallBack) {
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
		if atomic.LoadInt32(&s.closeFlag) == 1 {
			return
		}
		cn, acceptErr := s.listener.Accept()
		if acceptErr != nil {
			LOG.Error("[Start] accept error = %s\n", acceptErr.Error())
			continue
		}
		tcpCon, _ := cn.(*net.TCPConn)
		tcpCon.SetNoDelay(true)
		tcpCon.SetKeepAlive(true)

		index := atomic.AddUint32(&s.counter, 1)
		con := NewConnection(tcpCon, index, s.codecBuild(tcpCon, tcpCon))

		con.setConnectionCallBack(s.connectionCallBack)
		con.setReadCallBack(s.readCallBack)
		con.establish()

		go con.readLoop()
		go con.writeLoop()
		LOG.Info("[Start] accept a new connection name = %s, id = %d\n", con.name, con.id)
	}
}

func (s *Server) Close() {
	if atomic.CompareAndSwapInt32(&s.closeFlag, 0, 1) {
		LOG.Info("[Close] server closed")
		s.listener.Close()
	}
}
