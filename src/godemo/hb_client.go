package main

import (
	"flag"
	"heartbeat"
	. "logger"
	. "sock"
	"time"
)

func main() {
	address := flag.String("address", "127.0.0.1:8080", "server listen port")
	ms := flag.Int64("timeout", 5000, "read write channel timeout millsecond")
	num := flag.Int("num", 10, "the number of the connections")

	flag.Parse()
	LOG.SetLevel(0)
	client := heartbeat.NewHeartBeatClient(*address, LineCodecBuild, time.Duration((*ms)))
	for i := 0; i < *num; i++ {
		client.Connect()
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(60000 * time.Second)
}
