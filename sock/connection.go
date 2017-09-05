package sock

import (
	"fmt"
	. "github.com/elwin0214/gomemcached/util"
	"github.com/golang/glog"
	"net"
	"time"
)

type Connection struct {
	//base
	id        uint32
	name      string
	closeFlag AtomicInt32
	//io
	tcpConn *net.TCPConn
	codec   Codec
	flusher Flusher

	//callback
	connectionCallBack ConnectionCallBack

	//read
	readCallBack ReadCallBack
	readTimeout  time.Duration

	writeChan chan interface{}
}

func NewConnection(tcpConn *net.TCPConn, index uint32) *Connection {
	con := new(Connection)
	con.id = index
	con.name = fmt.Sprintf("%s-%d", tcpConn.RemoteAddr().String(), index)
	con.tcpConn = tcpConn
	con.writeChan = make(chan interface{}, 1024)
	con.closeFlag = 0
	return con
}

func (con *Connection) establish() {
	if nil != con.connectionCallBack {
		con.connectionCallBack(con)
	}
}
func (con *Connection) GetId() uint32 {
	return con.id
}

func (con *Connection) GetName() string {
	return con.name
}

func (con *Connection) SetNoDelay(flag bool) {
	con.tcpConn.SetNoDelay(flag)
}

func (con *Connection) SetKeepAlive(flag bool) {
	con.tcpConn.SetKeepAlive(flag)
}

func (con *Connection) SetCodec(codec Codec) {
	con.codec = codec
}

func (con *Connection) setConnectionCallBack(callback ConnectionCallBack) {
	con.connectionCallBack = callback
}

func (con *Connection) setReadCallBack(callback ReadCallBack) {
	con.readCallBack = callback
}

func (con *Connection) GetTcpConn() *net.TCPConn {
	return con.tcpConn
}

func (con *Connection) IsClosed() bool {
	return con.closeFlag.Get() == 1
}

func (con *Connection) Close() {
	if con.closeFlag.Cas(0, 1) {
		glog.Warningf("[Close] conn = %s\n", con.GetName())
		con.tcpConn.Close()
		if nil != con.connectionCallBack {
			con.connectionCallBack(con)
		}

		close(con.writeChan)
	} else {
		glog.Warningf("[Close] conn = %s closed\n", con.GetName())
	}
}

func (con *Connection) SetReadTimeout(timeoutMs time.Duration) {
	con.readTimeout = time.Millisecond * timeoutMs
}

func (con *Connection) readLoop() {

	for {
		if con.readTimeout > 0 {
			con.tcpConn.SetReadDeadline(time.Now().Add(con.readTimeout))
		}
		body, err := con.codec.Decode()
		if nil != err {
			glog.Warningf("[readLoop] conn = %s error = %s  goroute exit\n", con.GetName(), err.Error())
			con.Close()
			return
		}
		if nil != con.readCallBack {
			con.readCallBack(con, &Message{Id: con.id, Body: body})
		} else {
			glog.Infof("[readLoop] not exit read callback\n")
		}
	}
}

func (con *Connection) Send(msg interface{}) {
	if con.IsClosed() {
		glog.Warningf("[Send] conn closed\n")
		return
	}
	con.writeChan <- msg // if closed
}

func (con *Connection) writeLoop() {
	for {
		select {
		case msg, ok := <-con.writeChan:
			if !ok {
				glog.Errorf("[writeLoop] connection = %s, write channel is closed, goroute exit.\n", con.GetName())
				return
			}
			if con.IsClosed() {
				glog.Errorf("[writeLoop] connection = %s, connection is closed, goroute exit.\n", con.GetName())
				return
			}
			err := con.codec.Encode(msg)
			glog.Infof("[writeLoop] connection = %s write msg", con.GetName())
			if nil != err {
				con.Close()
				glog.Errorf("[writeLoop] connection = %s, error = %s, close conn, goroute exit.\n", con.GetName(), err.Error())
				return
			}
		}
	}
}
