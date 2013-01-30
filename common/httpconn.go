package common

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
)

type HTTPConn struct {
	conn net.Conn

	buf []byte
	err error
	req *http.Request
}

func NewHTTPConn(c net.Conn) (h *HTTPConn) {
	return &HTTPConn{conn: c}
}

func (c *HTTPConn) parseHeaders() (err error) {
	var buf [http.DefaultMaxHeaderBytes]byte

	n, err := c.conn.Read(buf[0:])
	if err != nil {
		log("H: error while reading", err)
		return err
	}
	log("H: bytes", n)
	c.buf = make([]byte, n)
	copy(c.buf, buf[0:n])

	c.req, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(c.buf[0:n])))
	if err != nil {
		log("H: error while parsing header")
		return err
	}
	return nil
}

func (c *HTTPConn) Host() string {
	if c.req != nil {
		return c.req.Host
	}

	err := c.parseHeaders()
	if err != nil {
		log("H: error", err)
		return ""
	}

	return c.req.Host
}

func (c *HTTPConn) Read(b []byte) (int, error) {
	// read from internal buffer
	if c.err != nil {
		return 0, c.err
	}
	if len(c.buf) != 0 {
		n := copy(b, c.buf)
		c.buf = c.buf[n:]
		log(fmt.Sprintf("copied: %d - remaining %d ", n, len(c.buf)))
		return n, nil
	}
	return c.conn.Read(b)
}
func (c *HTTPConn) Write(b []byte) (n int, err error) {
	return c.conn.Write(b)
}
func (c *HTTPConn) Close() error {
	return c.conn.Close()
}
