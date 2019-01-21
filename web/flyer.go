package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/tesujiro/smf3/data/db"
)

func (s *server) handleFlyers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

	fmt.Fprintf(w, string(flyerJson))
	return
}

func (s *server) handlePostFlyers(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		log.Printf("bad Content-Type!!")
		log.Printf(r.Header.Get("Content-Type"))
	}

	//To allocate slice for request body
	length, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		log.Printf("Content-Length failed!!")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Read body data to parse json
	body := make([]byte, length)
	_, err = r.Body.Read(body)
	if err != nil && err != io.EOF {
		log.Printf("read failed!!\n")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var flyer db.Flyer
	if err := json.Unmarshal(body, &flyer); err != nil {
		log.Printf("Request body unmarshaling  error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Println("flyer:%v\n", flyer)
	now := time.Now().Unix()
	flyer.ID = now //TODO: temporary
	flyer.StartAt = now
	flyer.EndAt = now + flyer.ValidPeriod
	if err := flyer.Set(); err != nil {
		log.Printf("Set Flyer error: (%v) flyer:%v\n", err, flyer)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	return
}
