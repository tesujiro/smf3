package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tesujiro/smf3/data/db"
)

func (s *server) hookNotifications() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodGet: // Only Get??
			s.hookNotification(w, r)
			return
		default:
			log.Printf("Http method error. Not Post nor Get : %v\n", r.Method)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

type WebhookRequest struct {
	Command string          `json:"command"`
	Group   string          `json:"group"`
	Detect  string          `json:"detect"`
	Hook    string          `json:"hook"`
	Key     string          `json:"key"`
	Id      string          `json:"id"`
	Time    string          `json:"time"`
	Object  json.RawMessage `json:"object"`
}

func (s *server) hookNotification(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received Webhook request!\n")

	length, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := make([]byte, length)
	length, err = r.Body.Read(body)
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var wr WebhookRequest
	err = json.Unmarshal(body[:length], &wr)
	if err != nil {
		log.Printf("Webhook request unmarshal error :%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//log.Printf("Request.Body: %s", wr)

	// FlyerID
	if len(strings.Split(wr.Hook, ":")) != 2 {
		log.Printf("Webhook request has wrong flyer id.:%s\n", wr.Key)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	flyerID := strings.Split(wr.Hook, ":")[1]
	flyer, err := db.GetFlyer(flyerID)
	if err != nil {
		log.Printf("DB get flyer error: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if flyer == nil {
		log.Printf("Flyer not found: flyer id.:%s\n", flyerID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//fmt.Printf("flyer=%#v\n", flyer)

	// GeoJson
	var feature db.GeoJsonFeature
	err = json.Unmarshal(wr.Object, &feature)
	if err != nil {
		log.Printf("Webhook request GeoJsonFeature unmarshal error :%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	loc, err := db.LocationFromFeature(feature)
	if err != nil {
		log.Printf("GeoJsonFeatue conversion error: %s feature(%s)\n", err, feature)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = s.CreateNotification(*flyer, *loc)
	if err != nil {
		log.Printf("Create notification error: %s geometry(%s)\n", err, feature.Geometry)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("webhook finished normally.\n")
}

func (s *server) CreateNotification(flyer db.Flyer, loc db.Location) error {
	now := time.Now().Unix()
	n := &db.Notification{
		ID:           fmt.Sprintf("%d:%d", flyer.ID, loc.ID),
		FlyerID:      int64(flyer.ID),
		UserID:       int64(loc.ID),
		Lat:          loc.Lat,
		Lon:          loc.Lon,
		DeliveryTime: now,
	}

	// check notification exists
	if exist, err := db.ExistNotification(n.ID); err != nil {
		return fmt.Errorf("DB get notification error: %s\n", err)
	} else if exist {
		return fmt.Errorf("notification already exists. ID:%v\n", n.ID)
	}

	// check stock
	if flyer.Stocked <= 0 {
		log.Printf("No flyer stock.\n")
		return nil
	}

	err := n.Set()
	if err != nil {
		return fmt.Errorf("DB set notification error: %s\n", err)
	}
	log.Printf("Set notification: %#v\n", n)

	flyer.Stocked--
	flyer.Delivered++
	err = flyer.Jset("properties.stocked", flyer.Stocked)
	if err != nil {
		return fmt.Errorf("DB set flyer stocked error: %s\n", err)
	}
	err = flyer.Jset("properties.delivered", flyer.Delivered)
	if err != nil {
		return fmt.Errorf("DB set flyer delivered error: %s\n", err)
	}
	return nil
}
