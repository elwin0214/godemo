package util

import (
	"testing"
)

func Test_List_Remove(t *testing.T) {
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

func Test_Clone(t *testing.T){
	l := NewList(8)
	for i := 0; i < 10; i++ {
		l.Append(i)
	}
	l2, n := l.Clone()
	t.Logf("len(l2) = %d n = %d\n", len(l2), n)
	for i := 0; i < 10; i++ {
		t.Logf("l[%d] = %d\n", i, l[i])
		if (l2[i] != i){
			t.Errorf("the element is %d\n", (l2)[i])
		}
	}
}
