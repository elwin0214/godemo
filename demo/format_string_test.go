package demo

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
)

func Benchmark_Sprintf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fmt.Sprintf("%d", 20)
	}
}

func Benchmark_FormatInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		strconv.FormatInt(20, 10)
	}
}

func Benchmark_Set_Sprintf(b *testing.B) {
	cmd := "set"
	key := "key"
	value := []byte("6666666666")
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	crlf := []byte("\r\n")
	for i := 0; i < b.N; i++ {
		s := fmt.Sprintf("%s %s %d %d %d\r\n", cmd, key, 0, 0, 10)
		buf.WriteString(s)
		buf.Write(value)
		buf.Write(crlf)
		buf.Reset()
	}
}

func Benchmark_Set_FormatInt(b *testing.B) {
	cmd := "set"
	key := "key"
	value := []byte("6666666666")
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	//crlf := []byte("\r\n")
	for i := 0; i < b.N; i++ {
		buf.WriteString(cmd)
		buf.WriteString(" ")
		buf.WriteString(key)
		buf.WriteString(" ")
		buf.WriteString(strconv.FormatInt(0, 10))
		buf.WriteString(" ")
		buf.WriteString(strconv.FormatInt(0, 10))
		buf.WriteString(" ")
		buf.WriteString(strconv.FormatInt(10, 10))
		buf.WriteString("\r\n")
		buf.Write(value)
		buf.WriteString("\r\n")
		buf.Reset()
	}
}
