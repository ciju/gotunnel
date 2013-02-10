package main

import (
	"./common"
	"flag"
	"fmt"
	"net"
	"os"
)

var (
	port         = flag.String("p", "", "port")
	externAddr   = flag.String("a", "", "the address to be used by the users")
	backproxyAdd = flag.String("x", "", "Proxy port to listen to")
	singleClient = flag.Bool("sc", true, "if no subdomain logic is needed.")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
}

func setupClient(eaddr, port string, adminc net.Conn, singleClient bool) {
	id := common.ReceiveSubRequest(adminc)

	log("Client: ", connStr(adminc), id)

	proxy := router.Register(adminc, id)

	clientURL, requestURL := proxy.Host(eaddr), net.JoinHostPort(eaddr, port)
	log("Client: --- sending", requestURL, clientURL)

	common.SendProxyInfo(adminc, requestURL, clientURL)
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

var router = common.NewTCPRouter(35000, 35002)

func main() {
	flag.Usage = Usage
	flag.Parse()

	if *port == "" || *backproxyAdd == "" || *externAddr == "" {
		flag.Usage()
		os.Exit(1)
	}

	go func() { // new clients
		backproxy, err := net.Listen("tcp", *backproxyAdd)
		if err != nil {
			fatal("Client: Coundn't start server to connect clients", err)
		}

		for {
			adminc, err := backproxy.Accept()
			if err != nil {
				fatal("Client: Problem accepting new client", err)
			}
			go setupClient(*externAddr, *port, adminc, *singleClient)
		}

	}()

	// new request
	server, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", *port))
	if server == nil {
		fatal("Request: cannot listen: %v", err)
	}

	for {
		conn, err := server.Accept()
		if err != nil {
			fatal("Request: failed to accept new request: ", err)
		}
		go fwdRequest(conn)
	}
}

func fatal(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "netfwd: %s\n", fmt.Sprintf(s, a))
	os.Exit(2)
}

func log(msg string, r ...interface{}) {
	fmt.Println(msg, r)
}

func connStr(conn net.Conn) string {
	return string(conn.LocalAddr().String()) + " <-> " + string(conn.RemoteAddr().String())
}
