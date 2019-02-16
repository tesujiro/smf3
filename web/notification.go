package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/tesujiro/smf3/data/db"
)

func (s *server) handleNotifications() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			log.Printf("bad Content-Type:%s", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		//log.Printf("Content-Type:%s\n", r.Header.Get("Content-Type"))

		switch r.Method {
		case http.MethodPost:
			s.handlePostNotifications(w, r)
			return
		case http.MethodGet:
			s.handleGetNotifications(w, r)
			return
		default:
			log.Printf("Http method error. Not Post nor Get : %v\n", r.Method)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (s *server) handleGetNotifications(w http.ResponseWriter, r *http.Request) {
	bounds := make(map[string]float64, 4)
	query := r.URL.Query()
	for k, v := range query {
		k = strings.ToLower(k)
		switch k {
		case "south", "west", "north", "east":
			if len(v) > 1 {
				log.Printf("Query parameter conversion error: %v has more than one values\n", k)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			f, err := strconv.ParseFloat(v[0], 64)
			if err != nil {
				log.Printf("Query parameter conversion error: %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if _, ok := bounds[k]; ok {
				log.Printf("Query parameter conversion error: %v duplicate key\n", k)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			bounds[k] = f
		default:
			log.Printf("Unknown query parameter: %v\n", k)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	if len(bounds) != 4 {
		log.Printf("Query parameter error: not all directions: %v\n", len(bounds))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3. notifications
	var notifications []db.GeoJsonFeature
	var notificationJson []byte
	var err error
	notifications, err = db.NotificationFeaturesWithinBounds(bounds["south"], bounds["west"], bounds["north"], bounds["east"])
	if err != nil {
		log.Printf("WithiLocation error: %v\n", err)
		return
	}
	notificationJson, err = json.Marshal(notifications)
	if err != nil {
		log.Printf("Notification Marshal error: %v\n", err)
		return
	}
	//fmt.Fprintf(w, "%s", notificationJson)

	fmt.Fprintf(w, string(notificationJson))
	return
}

func (s *server) handlePostNotifications(w http.ResponseWriter, r *http.Request) {
	return
}
