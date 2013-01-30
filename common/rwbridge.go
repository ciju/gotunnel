package common

import (
	"fmt"
	"io"
)

func copyFromTo(a, b io.ReadWriteCloser) {
	defer func() {
		log("closing connection")
		a.Close()
	}()
	io.Copy(a, b)
}

type RWBridge struct {
	src, dst io.ReadWriteCloser
}

func (p *RWBridge) Proxy() {
	// go copypaste(p.src, p.dst, false, "f")
	// go copypaste(p.dst, p.src, true, "b")
	go copyFromTo(p.src, p.dst)
	go copyFromTo(p.dst, p.src)
}

func NewRWBridge(src, dst io.ReadWriteCloser) (p *RWBridge) {
	b := &RWBridge{src: src, dst: dst}
	b.Proxy()
	return b
}

// // only the actual host/port client should be able to close a
// // connection. ex: Keep-Alive and websockets.

func copypaste(in, out io.ReadWriteCloser, close_in bool, msg string) {
	var buf [512]byte

	defer func() {
		if close_in {
			fmt.Println("eof closing connection")
			in.Close()
			out.Close()
		}
	}()

	for {
		n, err := in.Read(buf[0:])
		// on readerror, only bail if no other choice.
		if err == io.EOF {
			// fmt.Print(msg)
			// time.Sleep(1e9)
			fmt.Println("eof", msg)
			return
		}
		fmt.Println("-- read ", msg, n)
		if err != nil {
			fmt.Println("something wrong while copying in ot out ", msg)
			fmt.Println("error: ", err)
			return
		}
		// if n < 1 {
		// 	fmt.Println("nothign to read")
		// 	return
		// }

		fmt.Println("-- msg bytes", msg, string(buf[0:n]))

		_, err = out.Write(buf[0:n])
		if err != nil {
			fmt.Println("something wrong while copying out to in ")
			// fatal("something wrong while copying out to in", err)
			return
		}
	}
}
