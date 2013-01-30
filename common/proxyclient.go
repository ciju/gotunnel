package common

import (
	"io"
	"net"
)

// proxyclient should be able to request a subdomain. If unavailable,
// it should be associated with a random one. Send it back so it can
// show it to the remote operator.
type ProxyClient struct {
	AdminChannel io.ReadWriteCloser
	listenServer net.Listener
	host         string
	conn         chan net.Conn
}

func NewProxyClient(adminConn io.ReadWriteCloser, addr string) (p *ProxyClient, err error) {
	p = &ProxyClient{host: addr, AdminChannel: adminConn, conn: make(chan net.Conn)}

	// port 0 means get an available port.
	p.listenServer, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	go p.accept()

	return p, nil
}

func (p *ProxyClient) Addr() net.Addr {
	return p.listenServer.Addr()
}

func (p *ProxyClient) accept() {
	for {
		backconn, err := p.listenServer.Accept()
		log("New connection from backend: ", connStr(backconn))
		if err != nil {
			log("some problem %v", err)
		}

		p.conn <- backconn
	}
}

// func (p *ProxyClient) request() (net.Conn, error) {
// 	// send request for new connection on admin channel
// 	fmt.Fprintf(p.adminChannel, "new\n")

// 	// TODO: have a timeout also
// 	return <-p.conn, nil
// }

func (p *ProxyClient) Forward(c io.ReadWriteCloser) error {
	// bc, err := p.request()
	bc := <-p.conn
	log("Received new connection. Fowarding.. ")
	NewRWBridge(c, bc)
	return nil
}
