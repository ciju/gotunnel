package httpheadreader

import (
	l "../log"
	"bufio"
	"bytes"
	"net"
	"net/http"
)

type HTTPHeadReader struct {
	conn net.Conn

	buf []byte
	err error
	req *http.Request
}

func NewHTTPHeadReader(c net.Conn) (h *HTTPHeadReader) {
	return &HTTPHeadReader{conn: c}
}

func (c *HTTPHeadReader) parseHeaders() (err error) {
	var buf [http.DefaultMaxHeaderBytes]byte

	n, err := c.conn.Read(buf[0:])
	if err != nil {
		l.Log("H: error while reading", err)
		return err
	}
	l.Log("H: bytes", n)
	c.buf = make([]byte, n)
	copy(c.buf, buf[0:n])

	c.req, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(c.buf[0:n])))
	if err != nil {
		l.Log("H: error while parsing header")
		return err
	}
	return nil
}

func (c *HTTPHeadReader) Host() string {
	if c.req != nil {
		return c.req.Host
	}

	err := c.parseHeaders()
	if err != nil {
		l.Log("H: error", err)
		return ""
	}

	return c.req.Host
}

func (c *HTTPHeadReader) Read(b []byte) (int, error) {
	// read from internal buffer
	if c.err != nil {
		return 0, c.err
	}
	if len(c.buf) != 0 {
		n := copy(b, c.buf)
		c.buf = c.buf[n:]
		l.Log("copied: %d - remaining %d ", n, len(c.buf))
		return n, nil
	}
	return c.conn.Read(b)
}
func (c *HTTPHeadReader) Write(b []byte) (n int, err error) {
	return c.conn.Write(b)
}
func (c *HTTPHeadReader) Close() error {
	return c.conn.Close()
}
