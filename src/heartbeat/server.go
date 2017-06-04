package heartbeat

import (
	. "logger"
	. "sock"
	"time"
)

type HeartBeatServer struct {
	server        *Server
	connections   map[string]*Connection
	readTimeoutMs time.Duration
}

func NewHeartBeatServer(address string, codecBuild CodecBuild, readTimeoutMs time.Duration) *HeartBeatServer {
	s := new(HeartBeatServer)
	option := Option{NoDely: true, KeepAlive: true, ReadBufferSize: 1024, WriteBufferSize: 1024}

	s.server = NewServer(address, codecBuild, option)
	s.connections = make(map[string]*Connection)
	s.readTimeoutMs = readTimeoutMs
	s.server.OnConnect(func(con *Connection) {
		s.onConnect(con)
	})
	s.server.OnRead(func(con *Connection, msg *Message) {
		s.onRead(con, msg)
	})
	return s
}

func (hbs *HeartBeatServer) Listen() error {
	return hbs.server.Listen()
}

func (hbs *HeartBeatServer) Start() {
	hbs.server.Start()
}

func (hbs *HeartBeatServer) Close() {
	hbs.server.Close()
	for _, con := range hbs.connections {
		con.Close()
	}
}

func (hbs *HeartBeatServer) onConnect(con *Connection) {
	if con.IsClosed() {
		delete(hbs.connections, con.GetName())
	} else {
		con.SetReadTimeout(hbs.readTimeoutMs)
		hbs.connections[con.GetName()] = con
	}
}

func (hbs *HeartBeatServer) onRead(con *Connection, msg *Message) {
	body, _ := msg.Body.([]byte)
	LOG.Info("[onRead] connection = %s msg = %s\n", con.GetName(), string(body))
	con.Send(body)
}
