package main

import (
	"./common"
	"flag"
	"fmt"
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
func newClient(addr string, ac net.Conn, singleClient bool) (string, *common.ProxyClient) {
	fmt.Println("admin proxy: ", connStr(ac))

	h := common.ReceiveSubRequest(ac)
	if h == "" {
		h = newRandString()
	}

	host := addr
	if !singleClient {
		host = h + "." + host
	}

	p, err := common.NewProxyClient(ac, host)
	if err != nil {
		fatal("problem")
	}
	// send the channel info
	log("listening at %v", p.Addr())
	common.SendProxyInfo(ac, host, p.Addr().String())
	return host, p
}

var proxies = map[string]*common.ProxyClient{}

func listenForNewClients(addr string, backproxy net.Listener, singleClient bool) {
	for {
		ac, err := backproxy.Accept()
		if err != nil {
			log("Problem accepting new client")
			continue
		}

		host, proxy := newClient(addr, ac, singleClient)
		proxies[host] = proxy
	}
}

var (
	hosting      = flag.String("s", "", "server host and port")
	backproxyAdd = flag.String("p", "", "Proxy port to listen to")
	singleClient = flag.Bool("oc", true, "if no subdomain logic is needed.")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
}

func main() {
	// - single on multi mode (clients can request sub or not)
	// - port to run on (default)
	// - host to use, for connection resolution (probably host:port combination)

	flag.Usage = Usage
	flag.Parse()

	if *hosting == "" || *backproxyAdd == "" {
		flag.Usage()
		os.Exit(1)
	}

	serverAddr := *hosting
	clientListnerAddr := *backproxyAdd

	backproxy, err := net.Listen("tcp", clientListnerAddr)

	go listenForNewClients(serverAddr, backproxy, *singleClient)

	server, err := net.Listen("tcp", serverAddr)
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
		fmt.Println("connecting for host: ", hcon.Host())
		p, ok := proxies[hcon.Host()]
		if !ok {
			log("no clients registered or ", hcon.Host())
			hcon.Close()
			continue
		}
		common.SendConnRequest(p.AdminChannel)
		p.Forward(hcon)
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
