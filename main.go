package main

import (
	"flag"
	"fmt"
	"net/http"

	"stcache/cache"
)

func main() {
	var serverId = flag.String("s", "0", "serverId")
	var tcpAddr = flag.String("t", "127.0.0.1:9001", "tcpAddr")
	var joinAddr = flag.String("j", "", "joinAddr")
	var bootstrap = flag.Bool("b", false, "bootstrap")
	flag.Parse()
	fmt.Println("flag parse:", *serverId, *tcpAddr, *joinAddr, *bootstrap)

	opt := &cache.Options{
		ServerId:  *serverId,
		TcpAddr:   *tcpAddr,
		JoinAddr:  *joinAddr,
		Bootstrap: *bootstrap,
	}

	go func() {
		cache.Init(opt)
	}()
	http.HandleFunc("/get", cache.GetHandler)
	http.HandleFunc("/set", cache.SetHandler)
	http.HandleFunc("/join", cache.JoinHandler)
	http.HandleFunc("/loadtest", cache.StartLoadTest)


	port := ":800" + opt.ServerId
	http.ListenAndServe(port, nil)
}
