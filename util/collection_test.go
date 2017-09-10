package util

import (
	"testing"
)

func Test_Remove(t *testing.T) {
	l := NewList(8)
	for i := 0; i < 10; i++ {
		l.Append(i)
	}

	for i := 0; i < 10; i++ {
		if i % 2 == 0 {
			l.Remove(i, func(a, b interface{}) bool {
				ai, _ := a.(int)
				bi, _ := b.(int)
				return ai == bi

			})
		}
	}

	if l.Len() != 5 {
		t.Errorf("list.Len() is %d\n", l.Len())
	}
	for i := 0; i < 5; i++ {
		if l[i] != 2 * i + 1 {
			t.Errorf("the element is %d\n", l[i])
		}
	}
}

func Test_Clone(t *testing.T) {
	l := NewList(8)
	for i := 0; i < 10; i++ {
		l.Append(i)
	}
	l2, n := l.Clone()
	t.Logf("len(l2) = %d n = %d\n", len(l2), n)
	for i := 0; i < 10; i++ {
		t.Logf("l[%d] = %d\n", i, l[i])
		if (l2[i] != i) {
			t.Errorf("the element is %d\n", (l2)[i])
		}
	}
}

func Test_Remove_NoFound(t *testing.T) {
	l := NewList(8)
	for i := 0; i < 10; i++ {
		l.Append(i)
	}
	for i := 0; i < 10; i++ {
		index := l.Remove(i + 10, func(e1, e2 interface{}) bool{
			i1, _ := e1.(int)
			i2, _ := e2.(int)
			return i1 == i2
		})
		if index != -1 {
			t.Errorf("the index is %d\n", index)
		}
	}
	if 10 != l.Len() {
		t.Errorf("the length is %d\n", l.Len())
	}

}
func Test_Remove_oundary(t *testing.T) {
	l := NewList(8)
	for i := 0; i < 10; i++ {
		l.Append(i)
	}

	index := l.Remove(0, func(e1, e2 interface{}) bool{
		i1, _ := e1.(int)
		i2, _ := e2.(int)
		return i1 == i2
	})
	if index != 0 {
		t.Errorf("the index is %d\n", index)
	}
	if 9 != l.Len() {
		t.Errorf("the length is %d\n", l.Len())
	}
	index = l.Remove(9, func(e1, e2 interface{}) bool{
		i1, _ := e1.(int)
		i2, _ := e2.(int)
		return i1 == i2
	})
	if index != 8 {
		t.Errorf("the index is %d\n", index)
	}
	if 8 != l.Len() {
		t.Errorf("the length is %d\n", l.Len())
	}
}
