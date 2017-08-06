package main

import (
	"flag"
	. "logger"
	. "memcached"
	"os"
	"runtime"
	"runtime/pprof"
	. "sig"
)

func main() {

	address := flag.String("a", "127.0.0.1:8080", "server listen port")
	level := flag.Int("l", 2, "log level")
	flag.Parse()

	file, _ := os.Create("cpu.out")
	pprof.StartCPUProfile(file)
	LOG.SetHandler(NewStreamHandler(os.Stdout))
	LOG.SetLevel(*level)
	LOG.Warn("maxprocs = %d\n", runtime.GOMAXPROCS(0))
	s := NewMemcachedServer(*address, 4*1024)
	RegisterStopSignal(func() {
		pprof.StopCPUProfile()
		LOG.Info("close")
		s.Close()
	})
	s.Listen()
	s.Start()
}
