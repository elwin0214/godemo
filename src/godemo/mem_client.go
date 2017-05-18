package main

import (
	"flag"
	"fmt"
	. "logger"
	. "memcached"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

func main() {
	la := flag.String("la", "127.0.0.1:8080", "server listen port")
	pa := flag.String("pa", "127.0.0.1:8089", "server listen port")
	conns := flag.Int("conns", 1, "the number of tcp connection")
	vl := flag.Int("vl", 10, "the length of value")
	clients := flag.Int("cs", 10, "the number of clients")
	requests := flag.Int("reqs", 100, "the number of requests")
	level := flag.Int("level", 2, "log level")
	prof := flag.Bool("prof", true, "cpu profile")
	cpuprof := flag.String("cpuprof", "cpu.out", "cpu profile")
	flag.Parse()
	go func() {
		http.ListenAndServe(*pa, nil)
	}()
	LOG.SetHandler(NewStreamHandler(os.Stdout))
	LOG.SetLevel(*level)
	LOG.Warn("maxprocs = %d\n", runtime.GOMAXPROCS(0))

	if *prof {
		file, _ := os.Create(*cpuprof)
		pprof.StartCPUProfile(file)
		defer pprof.StopCPUProfile()
	}
	buf := make([]byte, 0, *vl)
	for j := 0; j < *vl; j++ {
		buf = append(buf, 'a')
	}
	value := string(buf)
	LOG.Info("value = %s\n", value)
	c := NewMemcachedClient(*la, *conns, 5000)
	c.Start()
	var wg sync.WaitGroup
	wg.Add(*clients)
	start := time.Now()
	for i := 0; i < *clients; i++ {
		go func(index int) {
			LOG.Info("start %d\n", index)

			defer wg.Done()
			for k := 0; k < *requests; k++ {
				key := fmt.Sprintf("%d_%d", index, k)

				r, err := c.Set(key, value)
				LOG.Info("key = %s value = %s result = %t err = %v\n", key, value, r, err)

			}
			LOG.Info("asd")
		}(i)
	}
	wg.Wait()
	end := time.Now()
	LOG.Warn("[main] client = %d requests = %d time = %dms\n", *clients, *requests, end.Sub(start)/1000/1000)
	time.Sleep(10000 * time.Millisecond)
	c.Close()
}
