package common

import (
	"io"
	"net"
	"strconv"
)

// proxyclient should be able to request a subdomain. If unavailable,
// it should be associated with a random one. Send it back so it can
// show it to the remote operator.
type ProxyClient struct {
	listenServer net.Listener
	conn         chan net.Conn
}

func NewProxyClient(port int) (p *ProxyClient, err error) {
	p = &ProxyClient{conn: make(chan net.Conn)}

	p.listenServer, err = net.Listen("tcp4", ":"+strconv.Itoa(port))
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

func (p *ProxyClient) Forward(c io.ReadWriteCloser) error {
	// bc, err := p.request()
	bc := <-p.conn
	log("Received new connection. Fowarding.. ")
	NewRWBridge(c, bc)
	return nil
}
