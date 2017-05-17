package util

import "sync/atomic"

type AtomicInt struct {
	value int32
}

func NewAtomicInt(value int32) *AtomicInt {
	return &AtomicInt{value: value}
}
func (a *AtomicInt) Cas(old int32, new int32) bool {
	return atomic.CompareAndSwapInt32(&a.value, old, new)
}

func (a *AtomicInt) Get() int32 {
	return atomic.LoadInt32(&a.value)
}
