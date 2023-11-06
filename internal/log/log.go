package log

import (
	"fmt"
	"io"
	"log"
)

var debug bool

func SetDebug(active bool) {
	debug = active
}

func Debug(msg string) {
	Debugf("%s\n", msg)
}

func Debugf(format string, v ...any) {
	if debug {
		log.Printf(fmt.Sprintf("DEBUG: %s", format), v...)
	}
}

func SetOutput(w io.Writer) {
	log.SetOutput(w)
}

func Printf(format string, v ...any) {
	log.Printf(format, v...)
}

func Println(v ...any) {
	log.Println(v...)
}
