package main

import (
	"./common"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

func connectRemoteProxy(addr string, times int) (conn net.Conn, err error) {
	conn, err = net.Dial("tcp", addr)
	if err != nil {
		if times > 0 {
			log("CMD: couldn't connect, trying again")
			time.Sleep(1e9)
			return connectRemoteProxy(addr, times-1)
		}
		return nil, err
	}
	return
}

func setupCommandChannel(addr string, req, quit chan bool, conn chan string) {
	backproxy, err := connectRemoteProxy(addr, 3)
	if err != nil {
		log("CMD: Couldn't connect to ", addr, "err: ", err)
		quit <- true
		return
	}
	defer backproxy.Close()

	common.SendSubRequest(backproxy, "")

	// the port to connect on
	serverat, conn_to, _ := common.ReceiveProxyInfo(backproxy)
	conn <- conn_to

	fmt.Println("go to ", serverat)

	for {
		req <- common.ReceiveConnRequest(backproxy)
	}
}

var (
	host   = flag.String("l", "", "host ip/name and port")
	remote = flag.String("r", "127.0.0.1:8001", "the remote gotunnel server host/ip:port")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
}

func main() {
	// - remote proxy address to listen to (default)
	// - localhost server port
	// - subdomain to request.

	flag.Usage = Usage
	flag.Parse()

	if *host == "" || *remote == "" {
		flag.Usage()
		os.Exit(1)
	}

	remoteProxy := *remote
	localServer := "127.0.0.1:" + *host

	lp, err := net.Dial("tcp", localServer)
	if err != nil {
		fmt.Fprintf(os.Stderr, `
 Local server not running. If your server is,
 running on some other port. Please mention it,
 in the options.

`)
		flag.Usage()
		os.Exit(1)
	}

	lp.Close()

	req, quit, conn := make(chan bool), make(chan bool), make(chan string)

	fmt.Printf("Setting Gotunnel server %s with local server on %s\n", *remote, *host)

	go setupCommandChannel(remoteProxy, req, quit, conn)

	remoteProxy = <-conn

	for {
		select {
		case <-req:
			fmt.Printf("New link b/w %s and %s\n", remoteProxy, localServer)
			rp, err := net.Dial("tcp", remoteProxy)
			if err != nil {
				log("problem", err)
				return
			}
			lp, err := net.Dial("tcp", localServer)
			if err != nil {
				log("problem", err)
				return
			}

			go common.NewRWBridge(rp, lp)
		case <-quit:
			return
		}
	}
}

func fatal(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "netfwd: %s\n", fmt.Sprintf(s, a))
	os.Exit(2)
}

func log(msg string, r ...interface{}) {
	fmt.Println(msg, r)
}
