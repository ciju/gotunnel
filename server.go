package main

import (
	"./httpheadreader"
	l "./log"
	proto "./protocol"
	"./tcprouter"
	"flag"
	"fmt"
	"net"
	"os"
)

// for isAlive
import (
	"io"
	"time"
)

// https://groups.google.com/d/topic/golang-nuts/e8sUeulwD3c/discussion
func isAlive(c net.Conn) bool {
	one := []byte{0}
	c.SetReadDeadline(time.Now())
	_, err := c.Read(one)
	if err == io.EOF {
		l.Log("%s detected closed LAN connection", c)
		c.Close()
		c = nil
		return false
	}

	c.SetReadDeadline(time.Time{})
	return true
}

func setupClient(eaddr, port string, adminc net.Conn) {
	id := proto.ReceiveSubRequest(adminc)

	l.Log("Client: asked for ", connStr(adminc), id)

	proxy := router.Register(adminc, id)

	requestURL, backendURL := proxy.FrontHost(eaddr, port), proxy.BackendHost(eaddr)
	l.Log("Client: --- sending %v %v", requestURL, backendURL)

	proto.SendProxyInfo(adminc, requestURL, backendURL)

	for {
		time.Sleep(2 * time.Second)
		if !isAlive(adminc) {
			router.Deregister(proxy)
			break
		}
	}
	l.Log("Client: closing backend connection")
}

func fwdRequest(conn net.Conn) {
	fmt.Println("Request: ", connStr(conn))
	hcon := httpheadreader.NewHTTPHeadReader(conn)

	if hcon.Host() == *externAddr || hcon.Host() == "www."+*externAddr {
		conn.Write([]byte(defaultMsg))
		conn.Close()
		return
	}

	p, ok := router.GetProxy(hcon.Host())
	if !ok {
		l.Log("Request: coundn't find proxy for", hcon.Host())
		conn.Write([]byte(fmt.Sprintf("Couldn't fine proxy for <%s>", hcon.Host())))
		conn.Close()
		return
	}

	proto.SendConnRequest(p.Admin)
	p.Proxy.Forward(hcon)
}

var router = tcprouter.NewTCPRouter(35000, 36000)
var defaultMsg = `
<html><body><style>body{background-color:lightGray;} h1{margin:0 auto;width:600px;padding:100px;text-align:center;} a{color:#4e4e4e;text-decoration:none;} h3{float:right;}</style><h1><a href="http://github.com/ciju/gotunnel">github.com/ciju/gotunnel</a></h1><h3>Sponsored by <a href="http://activesphere.com">ActiveSphere</a></h3></body></html>
`

var (
	port = flag.String("p", "32000", "Access the tunnel sites on this port.")
	// haproxy (or any other supporting WebSocket) can fwd the *80 traffic to the port above.
	externAddr   = flag.String("a", "localtunnel.net", "the address to be used by the users")
	backproxyAdd = flag.String("x", "0.0.0.0:34000", "Port for clients to connect to")
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
			l.Fatal("Client: Coundn't start server to connect clients", err)
		}

		for {
			adminc, err := backproxy.Accept()
			if err != nil {
				l.Fatal("Client: Problem accepting new client", err)
			}
			go setupClient(*externAddr, *port, adminc)
		}

	}()

	// new request
	server, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", *port))
	if server == nil {
		l.Fatal("Request: cannot listen: %v", err)
	}
	l.Log("Listening at: %s", *port)

	for {
		conn, err := server.Accept()
		if err != nil {
			l.Fatal("Request: failed to accept new request: ", err)
		}
		go fwdRequest(conn)
	}
}

func connStr(conn net.Conn) string {
	return string(conn.LocalAddr().String()) + " <-> " + string(conn.RemoteAddr().String())
}
