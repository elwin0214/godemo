package util

import (
	"sync"
	"testing"
)

func f(ai AtomicInt32, i int32) {
	old := ai.Get()
	ai.Cas(old, i)
}

func Test_Cas(t *testing.T) {
	var ai AtomicInt32
	ai = 0
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

}
