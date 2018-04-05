package termwin

import (
	"fmt"
	"io"
	"os"
)

var log io.Writer

func CreateLog(filename string) {
	file, err := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	log = file
}

func Logf(format string, args ...interface{}) {
	fmt.Fprintf(log, format, args...)
}

func Logln(s string) {
	fmt.Fprintln(log, s)
}
