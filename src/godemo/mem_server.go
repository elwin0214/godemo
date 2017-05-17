package main

import (
	"flag"
	. "logger"
	. "memcached"
	"net/http"
	_ "net/http/pprof"
	"os"
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

	LOG.SetHandler(NewStreamHandler(os.Stdout))
	LOG.SetLevel(*level)
	s := NewMemcachedServer(*la, NewMemcachedServerCodec)

	RegisterStopSignal(func() {
		LOG.Info("close")
		s.Close()

	})
	s.Listen()
	s.Start()
}
