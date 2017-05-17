package demo

import (
	"testing"
)

func Test_ReadClosedChannel(t *testing.T) {
	ch := make(chan int, 2)
	for i := 0; i < 2; i++ {
		ch <- i
	}
	close(ch)
	var j int
	ok := true
	for {
		j, ok = <-ch
		if !ok {
			t.Logf("channel is closed\n")
			break
		}
		t.Logf("%d\n", j)
	}
	_, ok = <-ch
	if !ok {
		t.Logf("channel is closed\n")
	}
}

func Test_WriteClosedChannel(t *testing.T) {

	defer func() {
		if err := recover(); err != nil {
			t.Logf("send to ch panic ", err)
		}
	}()

	ch := make(chan int, 2)
	for i := 0; i < 2; i++ {
		ch <- i
	}
	close(ch)
	for i := 0; i < 2; i++ {
		ch <- i
	}

}
