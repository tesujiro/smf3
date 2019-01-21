package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/tesujiro/smf3/data/db"
)

func (s *server) handleFlyers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			log.Printf("bad Content-Type!!")
			log.Printf(r.Header.Get("Content-Type"))
		}

		switch r.Method {
		case http.MethodPost:
			s.handlePostFlyers(w, r)
			return
		case http.MethodGet:
			s.handleGetFlyers(w, r)
			return
		default:
			log.Printf("Http method error. Not Post nor Get : %v\n", r.Method)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (s *server) handleGetFlyers(w http.ResponseWriter, r *http.Request) {
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

	var flyers []db.GeoJsonFeature
	var flyerJson []byte
	var err error
	now := time.Now().Unix()
	flyers, err = db.FlyerWithinBounds(bounds["south"], bounds["west"], bounds["north"], bounds["east"], "WHERE", "start", "-inf", now, "WHERE", "end", now, "+inf")
	if err != nil {
		log.Printf("WithiLocation error: %v\n", err)
		return
	}
	flyerJson, err = json.Marshal(flyers)
	if err != nil {
		log.Printf("Flyer Marshal error: %v\n", err)
		return
	}
	//fmt.Printf("flyerJson:%s\n", flyerJson)

	fmt.Fprintf(w, string(flyerJson))
	return
}

func (s *server) handlePostFlyers(w http.ResponseWriter, r *http.Request) {
	return
}
