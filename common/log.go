package common

import (
	"fmt"
	"net"
	"os"
)

func log(msg string, r ...interface{}) {
	fmt.Println(msg, r)
}

func fatal(msg string, r ...interface{}) {
	fmt.Println(msg, r)
	os.Exit(1)
}

func connStr(conn net.Conn) string {
	return string(conn.LocalAddr().String()) + " <-> " + string(conn.RemoteAddr().String())
}
