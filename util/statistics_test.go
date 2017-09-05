package util

import (
	"testing"
	"time"
)

func Test_Stat(t *testing.T) {
	s := NewStat(1000, 0, 10)
	s.Start()
	s.Collect(1)
	s.Collect(0)
	s.Collect(0)
	s.Collect(0)
	s.Collect(0)
	s.Collect(2)
	s.Collect(2)
	s.Collect(6)
	s.Collect(2)
	time.Sleep(1000 * time.Millisecond)
	s.Close()
	t.Log(s.View())
}
