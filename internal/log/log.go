package log

import (
	"fmt"
	"io"
	"log"
)

const (
	BOLD_BLACK   = "\033[30;1m"
	BOLD_RED     = "\033[31;1m"
	BOLD_GREEN   = "\033[32;1m"
	BOLD_YELLOW  = "\033[33;1m"
	BOLD_BLUE    = "\033[34;1m"
	BOLD_MAGENTA = "\033[35;1m"
	BOLD_CYAN    = "\033[36;1m"
	BOLD_WHITE   = "\033[37;1m"
	YELLOW       = "\033[0;33m"
	RESET        = "\033[0;m"
	CLEAR        = "\033[0K"
)

var debug bool
var quietLogging bool

func SetDebug(active bool) {
	debug = active
}

func SetQuietLogging(active bool) {
	quietLogging = active
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

func Debugln(v ...any) {
	Debugf("%s\n", v...)
}

func Debugf(format string, v ...any) {
	if debug {
		log.Printf(fmt.Sprintf("DEBUG: %s", format), v...)
	}
}

func Warnf(format string, v ...any) {
	log.Printf("%sWARN: %s%s", BOLD_YELLOW, fmt.Sprintf(format, v...), RESET)
}

func Warnln(v ...any) {
	log.Printf("%sWARN: %s%s", BOLD_YELLOW, fmt.Sprint(v...), RESET)
}

func ConditionalWarnf(format string, v ...any) {
	if !quietLogging {
		Warnf(format, v...)
	}
}

func ConditionalWarnln(v ...any) {
	if !quietLogging {
		Warnln(v...)
	}
}

func Errorf(format string, v ...any) {
	log.Printf("%sERROR: %s%s", BOLD_RED, fmt.Sprintf(format, v...), RESET)
}

func Errorln(v ...any) {
	log.Printf("%sERROR: %s%s", BOLD_RED, fmt.Sprint(v...), RESET)
}

func ConditionalErrorf(format string, v ...any) {
	if !quietLogging {
		Errorf(format, v...)
	}
}

func ConditionalErrorln(v ...any) {
	if !quietLogging {
		Errorln(v...)
	}
}

func Colorf(format string, v ...any) {
	log.Printf("%s%s%s", BOLD_MAGENTA, fmt.Sprintf(format, v...), RESET)
}

func Colorln(v ...any) {
	log.Printf("%s%s%s", BOLD_MAGENTA, fmt.Sprint(v...), RESET)
}

func ConditionalColorf(format string, v ...any) {
	if !quietLogging {
		Colorf(format, v...)
	}
}

func ConditionalColorln(v ...any) {
	if !quietLogging {
		Colorln(v...)
	}
}
