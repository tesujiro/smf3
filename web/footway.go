package main

import (
	"fmt"
	"log"
	"net/http"
)

func (s *server) handleFootway() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Footway Request:")

		w.WriteHeader(http.StatusOK)

		data, err := getFootway()
		if err != nil {
			log.Printf("Read json file failed!!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, string(data))

	}
}
