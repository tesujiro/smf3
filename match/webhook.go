package match

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

func CreateNotification(w http.ResponseWriter, r *http.Request) {
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

	c, err := feature.Geometry.GetCoordinatesObject()
	if err != nil {
		log.Printf("Geometry conversion error: %s geometry(%s)\n", err, feature.Geometry)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	point, ok := c.(*db.Point)
	if !ok {
		log.Printf("Coordinates conversion error: not point format:%s\n", feature.Geometry)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	userID := int64(feature.Properties["id"].(float64))
	now := time.Now().Unix()
	n := &db.Notification{
		ID:           fmt.Sprintf("%d:%d", flyer.ID, userID),
		FlyerID:      int64(flyer.ID),
		UserID:       int64(userID),
		Lat:          point[1],
		Lon:          point[0],
		DeliveryTime: now,
	}

	// check notification exists
	if exist, err := db.ExistNotification(n.ID); err != nil {
		log.Printf("DB get notification error: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if exist {
		log.Printf("notification already exists. ID:%v\n", n.ID)
		w.WriteHeader(http.StatusOK)
		return
	}

	// check stock
	if flyer.Stocked <= 0 {
		log.Printf("No flyer stock.\n")
		w.WriteHeader(http.StatusOK)
		return
	}

	err = n.Set()
	if err != nil {
		log.Printf("DB set notification error: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("Set notification: %#v\n", n)
	//n.StoreCache()

	flyer.Stocked--
	flyer.Delivered++
	err = flyer.Jset("properties.stocked", flyer.Stocked)
	if err != nil {
		log.Printf("DB set flyer stocked error: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = flyer.Jset("properties.delivered", flyer.Delivered)
	if err != nil {
		log.Printf("DB set flyer delivered error: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("webhook finished normally.\n")
}
