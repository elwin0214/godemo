package sock

import (
	. "logger"
	"net"
	. "util"
)

type Server struct {
	address            string
	listener           *net.TCPListener
	counter            uint32
	codecBuild         CodecBuild
	closeFlag          AtomicInt32
	option             Option
	connectionCallBack ConnectionCallBack
	readCallBack       ReadCallBack
}

func NewServer(address string, codecBuild CodecBuild, option Option) *Server {
	s := &Server{address: address}
	s.counter = 0
	s.closeFlag = 0
	s.codecBuild = codecBuild
	s.option = option
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
		cn, acceptErr := s.listener.Accept()
		if acceptErr != nil {
			LOG.Error("[Start] accept error = %s\n", acceptErr.Error())
			continue
		}
		tcpCon, _ := cn.(*net.TCPConn)
		tcpCon.SetNoDelay(s.option.NoDely)
		tcpCon.SetKeepAlive(s.option.KeepAlive)
		s.counter = s.counter + 1
		index := s.counter
		//writer := bufio.NewWriterSize(tcpCon, s.option.WriteBufferSize)
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
	if s.closeFlag.Cas(0, 1) {
		LOG.Info("[Close] server closed")
		s.listener.Close()
	}
}
