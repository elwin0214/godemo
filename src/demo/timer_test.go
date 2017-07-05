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

func Test_ExpireTimer(t *testing.T) {
	timer := time.NewTimer(1 * time.Second)
	time.Sleep(2 * time.Second)
	<-timer.C
	t.Logf("Timer has expired.")
}

func Test_ResetExpireTimer(t *testing.T) {
	timer := time.NewTimer(1 * time.Second)
	time.Sleep(2 * time.Second)
	timer.Reset(2 * time.Second)
	<-timer.C
	t.Logf("Timer has expired.")
	<-timer.C
	t.Logf("Timer has expired.")
	select {
	case <-timer.C:
		t.Logf("Timer has expired.")
	default:
		t.Logf("Timer has not expired.")
	}
}

func Test_StopExpireTimer(t *testing.T) {
	timer := time.NewTimer(2 * time.Second)
	timer.Stop()
	select {
	case <-timer.C:
		t.Logf("Timer has expired.")
	default:
		t.Logf("Timer has not expired.")
	}

}
