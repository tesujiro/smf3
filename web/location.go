package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/tesujiro/smf3/data/db"
)

func (s *server) handleLocations() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//log.Printf("LocationsAPI Received:")

		if r.Header.Get("Content-Type") != "application/json" {
			log.Printf("bad Content-Type!!")
			log.Printf(r.Header.Get("Content-Type"))
		}

		switch r.Method {
		case http.MethodPost:
			s.handlePostLocations(w, r)
			return
		case http.MethodGet:
			s.handleGetLocations(w, r)
			return
		default:
			log.Printf("Http method error. Not Post nor Get : %v\n", r.Method)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}
}

func (s *server) handleGetLocations(w http.ResponseWriter, r *http.Request) {
	bounds := make(map[string]float64, 4)
	query := r.URL.Query()
	for k, v := range query {
		//fmt.Printf("Query %v:%v\n", k, v)
		if len(v) > 1 {
			log.Printf("Query parameter conversion error: %v has more than one values\n", k)
			return
		}
		f, err := strconv.ParseFloat(v[0], 64)
		if err != nil {
			log.Printf("Query parameter conversion error: %v\n", err)
			return
		}
		bounds[k] = f
	}

	var locations []db.GeoJsonFeature
	var locationJson []byte
	var err error
	locations, err = db.LocationWithinBounds(bounds["south"], bounds["west"], bounds["north"], bounds["east"])
	if err != nil {
		log.Printf("WithiLocation error: %v\n", err)
		return
	}
	locationJson, err = json.Marshal(locations)
	if err != nil {
		log.Printf("Location Marshal error: %v\n", err)
		return
	}
	//fmt.Fprintf(w, "%s", locationJson)
	fmt.Fprintf(w, string(locationJson))
	return
}

func (s *server) handlePostLocations(w http.ResponseWriter, r *http.Request) {
	return
}