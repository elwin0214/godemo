package demo

import (
	"testing"
)

type Object struct {
	slice []string
	i     int
}

func Test_Interface(t *testing.T) {
	var val interface{} = nil
	if val != nil {
		t.Errorf("'val != nil' is true\n")
	}
	var val2 interface{} = (*interface{})(nil)
	if val2 == nil {
		t.Errorf("val2 == nil is %t\n", (val2 == nil))
	}

}
