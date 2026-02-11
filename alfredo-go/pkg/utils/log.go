package utils

import (
	"fmt"
	"os"
)

// Log writes messages to stderr, similar to the Python version
func Log(format string, args ...interface{}) {
	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	} else {
		fmt.Fprintln(os.Stderr, format)
	}
}
