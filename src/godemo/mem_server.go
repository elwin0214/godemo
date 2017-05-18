package main

import (
	"flag"
	. "logger"
	. "memcached"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	. "sig"
)

func main() {

	la := flag.String("la", "127.0.0.1:8080", "server listen port")
	pa := flag.String("pa", "127.0.0.1:8888", "server profile port")
	level := flag.Int("level", 2, "log level")

	flag.Parse()
	go func() {
		http.ListenAndServe(*pa, nil)
	}()

	file, _ := os.Create("cpu.out")
	pprof.StartCPUProfile(file)
	LOG.SetHandler(NewStreamHandler(os.Stdout))
	LOG.SetLevel(*level)
	LOG.Warn("maxprocs = %d\n", runtime.GOMAXPROCS(0))

	s := NewMemcachedServer(*la, NewMemcachedServerCodec)
	RegisterStopSignal(func() {
		pprof.StopCPUProfile()
		LOG.Info("close")
		s.Close()
	})
	s.Listen()
	s.Start()
}
