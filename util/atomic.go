package util

import "sync/atomic"

type AtomicInt32 int32

func (a *AtomicInt32) Cas(old int32, new int32) bool {
	return atomic.CompareAndSwapInt32((*int32)(a), old, new)
}

func (a *AtomicInt32) Set(v int32) {
	atomic.LoadInt32((*int32)(a))
}

func (a *AtomicInt32) Get() int32 {
	return atomic.LoadInt32((*int32)(a))
}

type AtomicInt64 int64

func (a *AtomicInt64) Cas(old int64, new int64) bool {
	return atomic.CompareAndSwapInt64((*int64)(a), old, new)
}

func (a *AtomicInt64) Set(v int64) {
	atomic.StoreInt64((*int64)(a), v)
}

func (a *AtomicInt64) Get() int64 {
	return atomic.LoadInt64((*int64)(a))
}

type AtomicBool int32

func NewAtomicBool(b bool) AtomicBool {
	if b {
		return 1
	} else {
		return 0
	}
}

func (a *AtomicBool) Cas(old, new bool) bool {
	var o, n int32
	if old {
		o = 1
	}
	if new {
		n = 1
	}
	return atomic.CompareAndSwapInt32((*int32)(a), o, n)
}

func (a *AtomicBool) Set(b bool) {
	var v int32
	if b {
		v = 1
	}
	atomic.StoreInt32((*int32)(a), v)
}

func (a *AtomicBool) Get() bool {
	return 1 == atomic.LoadInt32((*int32)(a))
}
