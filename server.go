package main

import (
	"./common"
	"flag"
	"fmt"
	"net"
	"os"
)

// TODO: time and again, send a ping to the client. And disconnect if no response.

func setupClient(eaddr, port string, adminc net.Conn) {
	id := common.ReceiveSubRequest(adminc)

	log("Client: asked for ", connStr(adminc), id)

	proxy := router.Register(adminc, id)

	requestURL, backendURL := proxy.FrontHost(eaddr, port), proxy.BackendHost(eaddr)
	log("Client: --- sending %v %v", requestURL, backendURL)

	common.SendProxyInfo(adminc, requestURL, backendURL)
}

func fwdRequest(conn net.Conn) {
	fmt.Println("Request: ", connStr(conn))
	hcon := common.NewHTTPConn(conn)

	p, ok := router.GetProxy(hcon.Host())
	if !ok {
		log("Request: coundn't find proxy for", hcon.Host())
		return
	}

	common.SendConnRequest(p.Admin)
	p.Proxy.Forward(hcon)
}

var router = common.NewTCPRouter(35000, 36000)

var (
	port         = flag.String("p", "", "port")
	externAddr   = flag.String("a", "", "the address to be used by the users")
	backproxyAdd = flag.String("x", "", "Proxy port to listen to")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = Usage
	flag.Parse()

	if *port == "" || *backproxyAdd == "" || *externAddr == "" {
		flag.Usage()
		os.Exit(1)
	}

	// new clients
	go func() {
		backproxy, err := net.Listen("tcp", *backproxyAdd)
		if err != nil {
			fatal("Client: Coundn't start server to connect clients", err)
		}

		for {
			adminc, err := backproxy.Accept()
			if err != nil {
				fatal("Client: Problem accepting new client", err)
			}
			go setupClient(*externAddr, *port, adminc)
		}

	}()

	// new request
	server, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", *port))
	if server == nil {
		fatal("Request: cannot listen: %v", err)
	}
	log("Listening at: %s", *port)

	for {
		conn, err := server.Accept()
		if err != nil {
			fatal("Request: failed to accept new request: ", err)
		}
		go fwdRequest(conn)
	}
}

// TODO: move these functions to a common place.
func fatal(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "goltunnel: %s\n", fmt.Sprintf(s, a...))
	os.Exit(2)
}

func log(msg string, r ...interface{}) {
	fmt.Println(fmt.Sprintf(msg, r...))
}

func connStr(conn net.Conn) string {
	return string(conn.LocalAddr().String()) + " <-> " + string(conn.RemoteAddr().String())
}
