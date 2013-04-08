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

func Info(msg string, r ...interface{}) {
	fmt.Printf("\033[1;34m%s\033[0m\n", fmt.Sprintf(msg, r...))
}
