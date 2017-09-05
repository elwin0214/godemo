package demo

import (
	"testing"
	"unsafe"
)

func Test_Empty_Struct(t *testing.T) {
	var s struct{}
	t.Logf("%d\n", unsafe.Sizeof(s))
}
