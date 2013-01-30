package common

import (
	"fmt"
	"net"
)

func log(msg string, r ...interface{}) {
	fmt.Println(msg, r)
}

func connStr(conn net.Conn) string {
	return string(conn.LocalAddr().String()) + " <-> " + string(conn.RemoteAddr().String())
}
