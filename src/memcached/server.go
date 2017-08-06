package memcached

import (
	. "logger"
	. "sock"
	"sync"
)

type MemcachedServer struct {
	server         *Server
	storage        *Storage
	readBufferSize int
	connections    map[string]*Connection
	mutex          sync.Mutex
}

func NewMemcachedServer(address string, readBufferSize int) *MemcachedServer {
	s := new(MemcachedServer)
	s.server = NewServer(address)
	s.storage = NewStorage()
	s.readBufferSize = readBufferSize
	s.connections = make(map[string]*Connection, 1024)
	return s
}
func (s *MemcachedServer) Listen() {
	s.server.OnRead(func(con *Connection, msg *Message) {
		s.onRead(con, msg)
	})
	s.server.OnConnect(func(con *Connection) {
		s.onConnect(con)
	})
	s.server.Listen()
}
func (s *MemcachedServer) Start() {
	go s.storage.Loop()
	s.server.Start()
}

func (s *MemcachedServer) onConnect(con *Connection) {
	//todo concurrent write
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if con.IsClosed() {
		delete(s.connections, con.GetName())
	} else {
		codec := NewMemcachedServerCodec(con.GetTcpConn(), con.GetTcpConn(), s.readBufferSize)
		con.SetCodec(codec)
		s.connections[con.GetName()] = con
	}
}

func (s *MemcachedServer) Close() {

	s.server.Close()
	s.storage.exit()

	s.mutex.Lock()
	conns := make([]*Connection, 0, len(s.connections))
	for _, con := range s.connections {
		conns = append(conns, con)
	}
	s.mutex.Unlock()

	for _, con := range conns {
		con.Close()
	}
}

func (s *MemcachedServer) onRead(con *Connection, msg *Message) {
	req, _ := msg.Body.(*MemRequest)
	LOG.Debug("[onRead] %v\n", req)
	if "" != req.Err {
		//con.Write([]byte(req.Err))
		con.Send(&MemResponse{Err: req.Err})
		return
	}
	resp := s.storage.Dispatch(req)
	con.Send(resp)
}
