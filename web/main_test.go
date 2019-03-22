package main

import (
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.SetFlags(log.Lmicroseconds)
	//log.SetOutput(ioutil.Discard)
	//debug.On()
	v := m.Run()
	os.Exit(v)
}
