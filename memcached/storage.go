package memcached

import (
	"context"
	"github.com/golang/glog"
	"strconv"
)

type Storage struct {
	data   map[string]*Item
	ch     chan chanReq
	ctx    context.Context
	cancel func()
}

func NewStorage() *Storage {
	s := new(Storage)
	s.data = make(map[string]*Item, 102400)
	s.ch = make(chan chanReq, 1024)
	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancel = cancel
	return s
}

type chanReq struct {
	request      *MemRequest
	chanResponse chan *MemResponse
}

func (s *Storage) Dispatch(request *MemRequest) *MemResponse {
	r := chanReq{
		request:      request,
		chanResponse: make(chan *MemResponse, 1),
	}
	s.ch <- r
	return <-r.chanResponse
}

func (s *Storage) Loop() {
	for {
		select {
		case cr := <-s.ch:
			s.handle(cr)
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Storage) handle(cr chanReq) {
	op := cr.request.Op
	switch op {
	case ADD:
		cr.chanResponse <- s.handleAdd(cr.request)
	case SET:
		cr.chanResponse <- s.handleSet(cr.request)
	case REPLACE:
		cr.chanResponse <- s.handleReplace(cr.request)
	case GET:
		resp := s.handleGet(cr.request)
		cr.chanResponse <- resp

	case DELETE:
		cr.chanResponse <- s.handleDelete(cr.request)

	case INCR:
		cr.chanResponse <- s.handleCount(cr.request)

	case DECR:
		cr.chanResponse <- s.handleCount(cr.request)

	default:
		glog.Errorf("%s\n", "unknown request")
	}
}

func (s *Storage) exit() {
	s.cancel()
}

func (s *Storage) T_exit() {
	s.exit()
}

func (s *Storage) handleGet(request *MemRequest) *MemResponse {
	item, ok := s.data[request.Key]
	if !ok {
		return &MemResponse{Op: request.Op, Result: false}
	}
	resp := new(MemResponse)
	resp.Op = request.Op
	resp.Result = true
	resp.Key = request.Key
	resp.Flags = item.flags
	resp.Exptime = item.exptime
	resp.Bytes = uint16(len(item.data))
	resp.Data = make([]byte, resp.Bytes, resp.Bytes)
	copy(resp.Data, item.data)
	return resp
}

func (s *Storage) handleAdd(request *MemRequest) *MemResponse {
	_, ok := s.data[request.Key]
	if ok {
		return &MemResponse{Op: request.Op, Result: false}
	}
	item := new(Item)
	item.key = request.Key
	item.flags = request.Flags
	item.exptime = request.Exptime
	item.data = request.Data
	s.data[item.key] = item
	return &MemResponse{Op: request.Op, Result: true}
}

func (s *Storage) handleSet(request *MemRequest) *MemResponse {
	item := new(Item)
	item.key = request.Key
	item.flags = request.Flags
	item.exptime = request.Exptime
	item.data = request.Data

	s.data[item.key] = item
	return &MemResponse{Op: request.Op, Result: true}
}

func (s *Storage) handleReplace(request *MemRequest) *MemResponse {
	_, ok := s.data[request.Key]
	if !ok {
		return &MemResponse{Op: request.Op, Result: false}
	}
	item := new(Item)
	item.key = request.Key
	item.flags = request.Flags
	item.exptime = request.Exptime
	item.data = request.Data

	s.data[item.key] = item
	return &MemResponse{Op: request.Op, Result: true}
}

func (s *Storage) handleDelete(request *MemRequest) *MemResponse {
	_, ok := s.data[request.Key]
	if !ok {
		return &MemResponse{Op: request.Op, Result: false}
	}
	delete(s.data, request.Key)
	return &MemResponse{Op: request.Op, Result: true}
}

func (s *Storage) handleCount(request *MemRequest) *MemResponse {
	item, ok := s.data[request.Key]
	if !ok {
		return &MemResponse{Op: request.Op, Result: false}
	}
	var value uint32
	if v, err := strconv.Atoi(string(item.data)); err == nil {
		value = uint32(v)
		if request.Op == INCR {
			var tmp int64
			tmp = int64(value) + int64(request.Value)
			if tmp > MAX_UINT32_VALUE {
				value = 0
			} else {
				value = uint32(tmp)
			}
		} else {
			if value <= request.Value {
				value = 0
			} else {
				value = value - request.Value
			}
		}
		item.data = []byte(strconv.FormatInt(int64(value), 10))
		return &MemResponse{Op: request.Op, Value: value, Result: true}
	} else {
		return &MemResponse{Op: request.Op, Value: value, Result: false, Err: "INVALID INTEGER\r\n"}
	}
}
