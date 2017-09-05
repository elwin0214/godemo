package util

import "sync/atomic"

type AtomicInt32 int32

func (a *AtomicInt32) Cas(old int32, new int32) bool {
	return atomic.CompareAndSwapInt32((*int32)(a), old, new)
}

func (a *AtomicInt32) Get() int32 {
	return atomic.LoadInt32((*int32)(a))
}
