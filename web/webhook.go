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
	"github.com/tesujiro/smf3/debug"
)

func (s *server) hookNotifications() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		//case http.MethodPost, http.MethodGet: // Only POST??
		case http.MethodPost:
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
	debug.Printf("[Webhook] Received Webhook request!\n")

	length, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		log.Printf("[Webhook] Webhook request Content-Length conversion error :%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := make([]byte, length)
	length, err = r.Body.Read(body)
	if err != nil && err != io.EOF {
		log.Printf("[Webhook] Webhook request read error :%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var wr WebhookRequest
	err = json.Unmarshal(body[:length], &wr)
	if err != nil {
		log.Printf("[Webhook] Webhook request unmarshal error :%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	debug.Printf("[Webhook] Request.Body: %s", wr)

	// FlyerID
	if len(strings.Split(wr.Hook, ":")) != 2 {
		log.Printf("[Webhook] Webhook request has wrong hook id. Key:%s ID:%v\n", wr.Key, wr.Hook)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	flyerID := strings.Split(wr.Hook, ":")[1]
	flyer, err := db.GetFlyer(flyerID)
	if err != nil {
		log.Printf("[Webhook] DB get flyer error: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if flyer == nil {
		log.Printf("[Webhook] Flyer not found: flyer id:%s\n", flyerID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// GeoJson
	var feature db.GeoJsonFeature
	//fmt.Printf("Object=%s\n", wr.Object)
	err = json.Unmarshal(wr.Object, &feature)
	if err != nil {
		log.Printf("[Webhook] Webhook request GeoJsonFeature unmarshal error :%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//fmt.Printf("feature=%#v\n", feature)

	loc, err := db.LocationFromFeature(feature)
	if err != nil {
		log.Printf("[Webhook] GeoJsonFeatue conversion error: %s feature(%s)\n", err, feature)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = s.CreateNotification(*flyer, *loc)
	if err != nil {
		log.Printf("[Webhook] Create notification error: %s geometry(%s)\n", err, feature.Geometry)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	debug.Printf("[Webhook] Webhook finished normally.\n")
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
		return fmt.Errorf("DB get notification error: %s", err)
	} else if exist {
		debug.Printf("[Webhook] notification already exists. ID:%v", n.ID)
		return nil
	}

	// check stock
	if flyer.Stocked <= 0 {
		debug.Printf("[Webhook] No flyer stock. flyer ID:%v", flyer.ID)
		return nil
	}

	err := n.Set()
	if err != nil {
		return fmt.Errorf("DB set notification error: %s", err)
	}
	debug.Printf("[Webhook] Set notification: %#v\n", n)

	flyer.Stocked--
	flyer.Delivered++
	err = flyer.Jset("properties.stocked", flyer.Stocked)
	if err != nil {
		return fmt.Errorf("DB set flyer stocked error: %s", err)
	}
	err = flyer.Jset("properties.delivered", flyer.Delivered)
	if err != nil {
		return fmt.Errorf("DB set flyer delivered error: %s", err)
	}
	return nil
}
