package main

import (
	"./common"
	"flag"
	"fmt"
	"net"
	"os"
)

// connect to server:
// - send the requested subdomain to server.
// - server replies back with a port to setup command channel on.
// - it also replies with the server address that users can access the site on.
func setupCommandChannel(addr, sub string, req, quit chan bool, conn chan string) {
	backproxy, err := net.Dial("tcp", addr)
	if err != nil {
		log("CMD: Couldn't connect to ", addr, "err: ", err)
		quit <- true
		return
	}
	defer backproxy.Close()

	common.SendSubRequest(backproxy, sub)

	// the port to connect on
	serverat, conn_to, _ := common.ReceiveProxyInfo(backproxy)
	conn <- conn_to

	fmt.Println("Connect to ", serverat)

	for {
		req <- common.ReceiveConnRequest(backproxy)
	}
}

func ensureServer(addr string) {
	lp, err := net.Dial("tcp", addr)
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
}

var (
	port      = flag.String("p", "", "port")
	subdomain = flag.String("sub", "", "request subdomain to serve on")
	remote    = flag.String("r", "localtunnel.net:34000", "the remote gotunnel server host/ip:port")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = Usage
	flag.Parse()

	if *port == "" || *remote == "" {
		flag.Usage()
		os.Exit(1)
	}

	localServer := net.JoinHostPort("127.0.0.1", *port)

	ensureServer(localServer)

	req, quit, conn := make(chan bool), make(chan bool), make(chan string)

	fmt.Printf("Setting Gotunnel server %s with local server on %s\n", *remote, *port)

	go setupCommandChannel(*remote, *subdomain, req, quit, conn)

	remoteProxy := <-conn
	log("remote proxy: %v", remoteProxy)

	for {
		select {
		case <-req:
			fmt.Printf("New link b/w %s and %s\n", remoteProxy, localServer)
			rp, err := net.Dial("tcp", remoteProxy)
			if err != nil {
				log("Coundn't connect to remote clientproxy", err)
				return
			}
			lp, err := net.Dial("tcp", localServer)
			if err != nil {
				log("Couldn't connect to localserver", err)
				return
			}

			go common.NewRWBridge(rp, lp)
		case <-quit:
			return
		}
	}
}

func fatal(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "oops: %s\n", fmt.Sprintf(s, a))
	os.Exit(2)
}

func log(msg string, r ...interface{}) {
	fmt.Printf(msg+"\n", r...)
}
