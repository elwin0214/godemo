package main

import (
	"flag"
	"heartbeat"
	. "logger"
	"net/http"
	_ "net/http/pprof"
	. "sock"
	"time"
)

func main() {

	LOG.SetLevel(LevelDebug)
	la := flag.String("listen address", "127.0.0.1:8080", "server listen port")
	ms := flag.Int64("timeout", 10000, "read timeout millsecond")
	pa := flag.String("profile address", "127.0.0.1:8888", "server profile port")

	flag.Parse()
	go func() {
		http.ListenAndServe(*pa, nil)
	}()
	server := heartbeat.NewHeartBeatServer(*la, LineCodecBuild, time.Duration(*ms))
	err := server.Listen()
	if err != nil {
		LOG.Error("server listenr error = %s\n", err.Error())
		return
	}
	server.Start()
}
