package main

import (
	"flag"
	. "github.com/elwin0214/gomemcached/memcached"
	. "github.com/elwin0214/gomemcached/sig"
	"github.com/golang/glog"
	"os"
	"runtime"
	"runtime/pprof"
)

func main() {

	address := flag.String("a", "0.0.0.0:8080", "server listen port")
	flag.Parse()

	file, _ := os.Create("cpu.out")
	pprof.StartCPUProfile(file)

	glog.Warningf("maxprocs = %d\n", runtime.GOMAXPROCS(0))
	s := NewMemcachedServer(*address, 4*1024)
	RegisterStopSignal(func() {
		pprof.StopCPUProfile()
		glog.Infof("close")
		s.Close()
	})
	s.Listen()
	s.Start()
}
