package main

import (
	"./common"
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

func main() {
	// if len(os.Args) != 3 { 
	//         fatal("usage: netfwd local remote") 
	// } 
	// localAddr := os.Args[1] 
	// remoteAddr := os.Args[2] 

	// - remote proxy address to listen to (default)
	// - localhost server port
	// - subdomain to request.

	remoteProxy := "127.0.0.1:8001"
	// localServer := "127.0.0.1:4001"
	localServer := os.Args[1]

	lp, err := net.Dial("tcp", localServer)
	if err != nil {
		fatal("local server not running")
	}
	lp.Close()

	req, quit, conn := make(chan bool), make(chan bool), make(chan string)

	go setupCommandChannel(remoteProxy, req, quit, conn)

	remoteProxy = <-conn

	for {
		select {
		case <-req:
			log("Connecting ", remoteProxy, localServer)
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
