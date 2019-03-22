package debug

import "log"

var debug bool

func On() {
	debug = true
}

func Off() {
	debug = false
}

func Println(a ...interface{}) {
	if debug {
		log.Println(a...)
	}
}

func Printf(format string, a ...interface{}) {
	if debug {
		log.Printf(format, a...)
	}
}
