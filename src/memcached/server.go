package memcached

import (
	. "logger"
	. "sock"
)

type MemcachedServer struct {
	server      *Server
	storage     *Storage
	connections map[string]*Connection
}

func NewMemcachedServer(address string, codecBuild CodecBuild) *MemcachedServer {
	s := new(MemcachedServer)
	s.server = NewServer(address, codecBuild)
	s.storage = NewStorage()
	s.connections = make(map[string]*Connection, 1024)
	return s
}
func (s *MemcachedServer) Listen() {
	s.server.SetReadCallBack(func(con *Connection, msg *Message) {
		s.onRead(con, msg)
	})
	s.server.SetConnectionCallBack(func(con *Connection) {
		s.onConnection(con)
	})
	s.server.Listen()
}
func (s *MemcachedServer) Start() {
	go s.storage.Loop()
	s.server.Start()
}

func (s *MemcachedServer) onConnection(con *Connection) {
	if con.IsClosed() {
		delete(s.connections, con.GetName())
	} else {
		s.connections[con.GetName()] = con
	}
}

func (s *MemcachedServer) Close() {
	s.server.Close()
	s.storage.exit()
	for _, con := range s.connections {
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
