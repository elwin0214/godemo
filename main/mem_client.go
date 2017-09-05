package main

import (
	"flag"
	"fmt"
	. "github.com/elwin0214/gomemcached/memcached"
	. "github.com/elwin0214/gomemcached/util"
	"github.com/golang/glog"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

func main() {
	address := flag.String("a", "127.0.0.1:8080", "server listen port")
	conns := flag.Int("t", 1, "the number of tcp connection")
	vl := flag.Int("vl", 10, "the length of value")
	clients := flag.Int("c", 10, "the number of clients")
	requests := flag.Int("r", 100, "the number of requests")
	flag.Parse()
	glog.Warningf("maxprocs = %d\n", runtime.GOMAXPROCS(0))

	cf, _ := os.Create("cpu.out")
	pprof.StartCPUProfile(cf)
	defer pprof.StopCPUProfile()

	buf := make([]byte, 0, *vl)
	for j := 0; j < *vl; j++ {
		buf = append(buf, 'a')
	}
	value := string(buf)
	glog.Infof("value = %s\n", value)
	c := NewMemcachedClient(*address, *conns, 5000)
	c.Start()
	var wg sync.WaitGroup
	wg.Add(*clients)
	stat := NewStat(1024*1024, 0, 100)
	stat.Start()
	start := time.Now()
	for i := 0; i < *clients; i++ {
		go func(index int) {
			glog.Infof("start %d\n", index)

			defer wg.Done()
			for k := 0; k < *requests / *clients; k++ {
				key := fmt.Sprintf("%d_%d", index, k)
				s := time.Now()
				r, err := c.Set(key, value)
				e := time.Now()
				elasp := int(e.Sub(s) / 1000 / 1000)
				stat.Collect(elasp)
				glog.Infof("key = %s value = %s result = %t err = %v elaspe = %d\n", key, value, r, err, elasp)

			}
		}(i)
	}
	wg.Wait()
	end := time.Now()
	var qps float64
	qps = float64(*requests) * 1.0 * 1000 * 1000 / (float64(end.Sub(start)*1.0) / 1000)
	glog.Warningf("[main] clients = %d reqs = %d time = %dms qps = %f\n", *clients, *requests, end.Sub(start)/1000/1000, qps)
	time.Sleep(1000 * time.Millisecond)

	stat.Close()
	glog.Warningf(stat.View())

	time.Sleep(10000 * time.Millisecond)
	c.Close()
}
