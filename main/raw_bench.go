package main

import (
	"flag"
	"github.com/golang/glog"
	"net"
	"time"
)

func server(address string, bufSize int) {

	ln, err := net.Listen("tcp", address)
	if err != nil {
		glog.Errorf("error = %s\n", err.Error())
		return
	}
	for {
		cn, ae := ln.Accept()
		if nil != ae {
			glog.Errorf("error = %s\n", ae.Error())
			return
		}
		sum := 0
		start := time.Now()

		for {
			buf := make([]byte, bufSize, bufSize)
			n, re := cn.Read(buf)
			if nil != re {
				glog.Errorf("error = %s\n", re.Error())
				break
			}
			sum = sum + n
		}
		end := time.Now()
		glog.Errorf("[main] sum = %dM time = %dms\n", sum/1024/1024, end.Sub(start)/1000/1000)
	}
}

func client(address string, sum int, bufSize int, ch chan bool) {
	cn, err := net.Dial("tcp", address)
	if err != nil {
		glog.Errorf("error = %s\n", err.Error())
		return
	}
	start := time.Now()
	buf := make([]byte, bufSize, bufSize)

	n := 0
	for {
		_, we := cn.Write(buf)
		if we != nil {
			break
		}
		n++
		if n >= sum {
			break
		}
	}
	end := time.Now()
	glog.Errorf("[main] sum = %dM time = %dms\n", sum*bufSize/1024/1024, end.Sub(start)/1000/1000)
	cn.Close()
	ch <- true
}

func main() {
	la := flag.String("la", "127.0.0.1:8080", "server listen port")
	num := flag.Int("num", 10, "the numbers of requests")
	mode := flag.String("mode", "s", "server")
	bs := flag.Int("bs", 4096, "buf size")
	flag.Parse()
	ch := make(chan bool, 1)
	if *mode == "s" {
		go server(*la, *bs)
		<-ch
	} else {
		go client(*la, *num, *bs, ch)
		<-ch
		time.Sleep(1000 * time.Millisecond)
	}
}
