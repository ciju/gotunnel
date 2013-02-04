package main

import (
	"./common"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
)

const (
	chars        = "abcdefghiklmnopqrstuvwxyz"
	subdomainLen = 4
)

// client request for a string, if already taken, get a new one. else
// use the one asked by client.
func newRandString() string {
	var str [subdomainLen]byte
	// rand.Seed(time.Now().Unix())
	for i := 0; i < subdomainLen; i++ {
		rnum := rand.Intn(len(chars))
		str[i] = chars[rnum]
	}
	return string(str[:])
}

// singleClient thingi
func newClient(eaddr string, adminc net.Conn, singleClient bool) (string, *common.ProxyClient) {
	fmt.Println("admin proxy: ", connStr(adminc))

	h := common.ReceiveSubRequest(adminc)
	if h == "" {
		h = newRandString()
	}

	host, _, err := net.SplitHostPort(eaddr)
	if err != nil {
		fatal("couldn't split eaddr", eaddr)
	}

	if !singleClient {
		host = h + "." + host
	}

	prox, err := common.NewProxyClient(35000) // TODO: remove this param
	if err != nil {
		fatal("problem", err)
	}
	// send the channel info
	log("listening at %v", prox.Addr())
	_, p, err := net.SplitHostPort(prox.Addr().String())
	if err != nil {
		fatal("coudn't split add", prox.Addr().String())
	}
	common.SendProxyInfo(adminc, eaddr, net.JoinHostPort(host, p))
	fmt.Println("--- sending", eaddr, net.JoinHostPort(host, p))
	return host, prox
}

var proxies = map[string]*Proxy{}

func listenForNewClients(addr string, backproxy net.Listener, singleClient bool) {
	for {
		ac, err := backproxy.Accept()
		if err != nil {
			fmt.Println("Problem accepting new client", err)
			os.Exit(1)
		}

		fmt.Println("new client")
		host, proxy := newClient(addr, ac, singleClient)
		proxies[host] = &Proxy{proxy: proxy, admin: ac}
		fmt.Println("[proxies] saving for: ", host, proxy.Addr().String())
	}
}

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

type Proxy struct {
	proxy *common.ProxyClient
	admin io.ReadWriteCloser
}

// serving port, external host
func main() {
	// The externally visible dns/ip of server.
	// internal ip of server.

	// port to listen users requests. (make a url out of
	// it to send to the client, to show on console)
	// probably get the host also from options.

	// if serving single client
	// - no subdomains: ignore the subdomain request. and send the server address.

	// otherwise check and allocate subdomain and use it to serve requests.

	// externAddr: used by the user, and by the server to
	// figure the client to fwd to request to.

	// send externAddr to client, and route connections based on (subdomain if present +) externAddr 

	flag.Usage = Usage
	flag.Parse()

	if *port == "" || *backproxyAdd == "" || *externAddr == "" {
		flag.Usage()
		os.Exit(1)
	}

	serverPort := *port
	backproxy, err := net.Listen("tcp", *backproxyAdd)

	fmt.Println("-----", *externAddr, net.JoinHostPort(*externAddr, serverPort))
	go listenForNewClients(net.JoinHostPort(*externAddr, serverPort), backproxy, *singleClient)

	server, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", serverPort))
	if server == nil {
		fatal("cannot listen: %v", err)
	}

	// front server
	for {
		conn, err := server.Accept()
		if conn == nil {
			fatal("accept failed: %v", err)
		}
		fmt.Println("Request: ", connStr(conn))
		// figure out the host the connection is
		// trying to connect to.

		hcon := common.NewHTTPConn(conn)
		h, _, err := net.SplitHostPort(hcon.Host())
		if err != nil {
			fatal("split didnt work", err)
		}
		fmt.Println("[proxies] fwd request to client for: ", h)

		p, ok := proxies[h]
		if !ok {
			log("no clients registered or ", hcon.Host())
			hcon.Close()
			continue
		}
		common.SendConnRequest(p.admin)
		p.proxy.Forward(hcon)
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
