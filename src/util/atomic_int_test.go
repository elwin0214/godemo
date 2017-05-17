package util

import (
	"sync"
	"testing"
	. "util"
)

func f(ai AtomicInt, i int32) {
	old := ai.Get()
	ai.Cas(old, i)
}

func Test_Cas(t *testing.T) {
	ai := NewAtomicInt(0)
	bi := ai

	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			for {
				v := ai.Get()
				r := ai.Cas(v, v+1)
				if r {
					return
				}
			}
		}()
	}

	wg.Wait()

	if 100 != ai.Get() {
		t.Errorf("atomic int is %d\n", ai.Get())
	}
	if 100 != bi.Get() {
		t.Errorf("atomic int is %d\n", bi.Get())
	}
	f(*bi, 2)
	if 100 != bi.Get() {
		t.Errorf("atomic int is %d\n", bi.Get())
	}
	ci := *bi
	bi.Cas(100, 1)
	if 100 != ci.Get() {
		t.Errorf("atomic int is %d\n", ci.Get())
	}
}
