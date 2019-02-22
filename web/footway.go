package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func (s *server) handleFootway() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Footway Request:")

		data, err := getFootway()
		if err != nil {
			log.Printf("Read json file failed: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, string(data))
	}
}

func getFootway() ([]byte, error) {
	path := "../data/osm/ways_on_browser.json"
	return ioutil.ReadFile(path)
}
