package demo

import (
	"testing"
	"time"
)

type A struct{}

func Test_Timer(t *testing.T) {
	timer := time.NewTimer(30 * time.Second)
	go func() {
		<-timer.C
		t.Logf("Timer has expired.")
	}()
	timer.Reset(0 * time.Second)
	time.Sleep(2 * time.Second)
}
