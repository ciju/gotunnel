package main

import (
	"flag"
	"fmt"
	"os"
)

import (
	"./gtclient"
	l "./log"
	"github.com/ciju/vercheck"
)

var (
	port         = flag.String("p", "", "port")
	subdomain    = flag.String("sub", "", "request subdomain to serve on")
	remote       = flag.String("r", "localtunnel.net:34000", "the remote gotunnel server host/ip:port")
	skipVerCheck = flag.Bool("sc", false, "Skip version check")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = Usage
	flag.Parse()

	if *port == "" || *remote == "" {
		flag.Usage()
		os.Exit(1)
	}

	if !*skipVerCheck {
		if vercheck.HasMinorUpdate(
			"https://raw.github.com/ciju/gotunnel/master/VERSION",
			"./VERSION",
		) {
			l.Info("\nNew version of Gotunnel is available. Please update your code and run again. Or start with option -sc to continue with this version.\n")
			os.Exit(0)
		}
	}

	servInfo := make(chan string)

	go func() {
		serverat := <-servInfo
		fmt.Printf("Your site should be available at: \033[1;34m%s\033[0m\n", serverat)
	}()

	if !gtclient.SetupClient(*port, *remote, *subdomain, servInfo) {
		flag.Usage()
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
