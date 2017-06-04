package demo

import (
	"fmt"
	"testing"
	"time"
)

func Test_Map_1024_1204_string(t *testing.T) {

	value := "1234567"
	m := make(map[string]string, 1024*1024)
	start := time.Now()
	for index := 0; index < 1024*1024; index++ {
		key := fmt.Sprintf("key-%d", index)
		m[key] = value
	}
	end := time.Now()
	t.Logf("map/1024*1024/1024*1024  cost %d us\n", end.Sub(start)/1000)
}

func Test_Map_0_string(t *testing.T) {

	value := "1234567"
	m := make(map[string]string, 0)
	start := time.Now()
	for index := 0; index < 1024*1024; index++ {
		key := fmt.Sprintf("key-%d", index)
		m[key] = value
	}
	end := time.Now()
	t.Logf("map/0/1024*1024  cost %d us\n", end.Sub(start)/1000)
}

func Test_Map_1024_1204_int(t *testing.T) {
	m := make(map[int]int, 1024*1024)
	start := time.Now()
	for index := 0; index < 1024*1024; index++ {
		m[index] = index
	}
	end := time.Now()
	t.Logf("map/1024*1024/1024*1024  cost %d us\n", end.Sub(start)/1000)
}

func Test_Map_0_int(t *testing.T) {
	m := make(map[int]int)
	start := time.Now()
	for index := 0; index < 1024*1024; index++ {
		m[index] = index
	}
	end := time.Now()
	t.Logf("map/0/1024*1024  cost %d us\n", end.Sub(start)/1000)
}
