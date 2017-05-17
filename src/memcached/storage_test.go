package memcached

import (
	. "logger"
	. "memcached"
	"testing"
)

func Test_Storage(t *testing.T) {
	LOG.SetLevel(LevelDebug)
	s := NewStorage()
	req := new(MemRequest)
	req.Op = GET
	req.Key = "a"
	go func() {
		s.Loop()
	}()
	s.Dispatch(req)
	s.Dispatch(req)
}
