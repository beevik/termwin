package termwin

import (
	"fmt"
	"io"
	"os"
)

var log io.Writer

// CreateLog creates a new file for log output.
func CreateLog(filename string) {
	file, err := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	log = file
}

// Logf logs a formatted line of text to the log output file.
func Logf(format string, args ...interface{}) {
	fmt.Fprintf(log, format, args...)
}

// Logln logs a newline-terminated string to the log output file.
func Logln(s string) {
	fmt.Fprintln(log, s)
}
