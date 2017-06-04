package demo

import (
	"sync"
	"testing"
)

type Object struct {
	buffer []byte
	value  int
}

func Benchmark_Alloc(b *testing.B) {
	sum := 0
	for i := 0; i < b.N; i++ {
		o := &Object{value: 1}
		sum = sum + o.value
	}
}

func Benchmark_PoolAlloc(b *testing.B) {
	p := &sync.Pool{
		New: func() interface{} {
			return &Object{value: 1}
		},
	}
	sum := 0
	for i := 0; i < b.N; i++ {
		o := p.Get().(*Object)
		sum = sum + o.value
		p.Put(o)
	}
}
