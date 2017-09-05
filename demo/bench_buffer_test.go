package demo

import (
	"bytes"
	"testing"
)

func Benchmark_Buffer_WriteString(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	s := "abc"
	for i := 0; i < b.N; i++ {
		buf.WriteString(s)
		buf.Reset()
	}
}

func Benchmark_Buffer_Write(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	s := []byte("abc")
	for i := 0; i < b.N; i++ {
		buf.Write((s))
		buf.Reset()
	}
}
