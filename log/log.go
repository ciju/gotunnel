package log

import (
	"fmt"
	"os"
)

func Fatal(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "goltunnel: %s\n", fmt.Sprintf(s, a...))
	os.Exit(2)
}

func Log(msg string, r ...interface{}) {
	fmt.Println(fmt.Sprintf(msg, r...))
}
