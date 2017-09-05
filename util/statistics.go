package util

import (
	"bytes"
	"fmt"
	"sync"
)

type Stat struct {
	ch     chan int
	stopCh chan bool
	data   map[int]int
	start  int
	end    int
	step   float32
	mutex  sync.Mutex
}

var BelowRangeIdx = -1
var OverRangeIdx = -2

func NewStat(size int, start int, end int) *Stat {
	s := new(Stat)
	s.ch = make(chan int, size)
	s.stopCh = make(chan bool, 1)
	s.data = make(map[int]int, 10)
	if start >= end {
		panic("start is greater than end")
	}
	s.start = start
	s.end = end
	s.step = float32(end-start) / 10
	return s
}

func (s *Stat) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data[BelowRangeIdx] = 0
	s.data[OverRangeIdx] = 0
	for idx := 0; idx < 10; idx++ {
		s.data[idx] = 0
	}
	go s.loop()
}

func (s *Stat) Close() {
	s.stopCh <- true
}

func (s *Stat) View() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	for idx := 0; idx < 10; idx++ {
		l := float32(float32(s.start) + float32(idx)*s.step)
		r := float32(float32(s.start) + float32(idx+1)*s.step)
		buffer.WriteString(fmt.Sprintf("[%f - %f] %d\n", l, r, s.data[idx]))
	}
	buffer.WriteString(fmt.Sprintf("[BelowRange] %d\n", s.data[BelowRangeIdx]))
	buffer.WriteString(fmt.Sprintf("[OverRange] %d\n", s.data[OverRangeIdx]))
	return buffer.String()
}

func (s *Stat) Collect(n int) {
	s.ch <- n
}

func (s *Stat) index(i int) int {
	if i < s.start {
		return BelowRangeIdx
	}
	if i >= s.end {
		return OverRangeIdx
	}
	n := int(float32(i-s.start) / s.step)
	return n
}

func (s *Stat) loop() {
	for {
		select {
		case i := <-s.ch:
			s.mutex.Lock()
			idx := s.index(i)
			v, ok := s.data[idx]
			if ok {
				s.data[idx] = v + 1
			} else {
				s.data[idx] = 1
			}
			s.mutex.Unlock()
		case <-s.stopCh:
			return
		}
	}
}
