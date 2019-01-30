package db

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Notification struct {
	ID           int64
	FlyerID      int64
	UserID       int64
	Lat          float64
	Lon          float64
	DeliveryTime int64
}

type notifCacheKey struct {
	FlyerID int64
	UserID  int64
}

var notifCache map[notifCacheKey]interface{}

func init() {
	notifCache = make(map[notifCacheKey]interface{})
}

var currentNotificationID int64 = 0

func NewNotificationID() int64 {
	currentNotificationID++
	return currentNotificationID
}

func (n *Notification) geoJson() (string, error) {
	feature := &GeoJsonFeature{
		Type: "Feature",
		Geometry: &Geometry{
			Type: "Point",
			//Coordinates: [2]float64{n.Lon, n.Lat},
			Coordinates: []byte(fmt.Sprintf("[%v,%v]", n.Lon, n.Lat)),
		},
		Properties: map[string]interface{}{
			"id":           n.ID,
			"flyerId":      n.FlyerID,
			"userId":       n.UserID,
			"deliveryTime": n.DeliveryTime,
		},
	}

	json, err := json.Marshal(feature)
	if err != nil {
		log.Printf("Notification feature marshal error: %v\n", err)
		return "", err
	}
	return string(json), nil
}

func GetNotification(id string) (*Notification, error) {
	// Connect Tile38
	c := pool.Get()
	defer c.Close()

	b, err := db_get(c, "notification", id)
	if err != nil {
		log.Fatalf("GET DB error: %v\n", err)
		return nil, err
	}
	if b == nil {
		return nil, nil
	}

	var n Notification
	err = json.Unmarshal(b, &n)
	if err != nil {
		log.Fatalf("Unmarshal error: %v\n", err)
		return nil, err
	}

	return &n, nil
}

func (n *Notification) OnCache() bool {
	if _, ok := notifCache[notifCacheKey{n.FlyerID, n.UserID}]; ok {
		return true
	} else {
		return false
	}
}

// TODO: Cache Garbage Collection
// remove cache about invalid flyerid

func (n *Notification) StoreCache() {
	notifCache[notifCacheKey{n.FlyerID, n.UserID}] = nil //interface{}{}
}

func (n *Notification) Set() error {
	// Connect Tile38
	c := pool.Get()
	defer c.Close()

	if json, err := n.geoJson(); err != nil {
		fmt.Printf("notification.geoJson() error: %v\n", err)
		return err
	} else {
		//fmt.Printf("GeoJSON:%v\n", json)
		err = db_set_json(c, "notification", fmt.Sprintf("%v", n.ID), json, "FIELD", "time", n.DeliveryTime, "NX") // NX: IF NOT EXISTS
		if err != nil {
			log.Fatalf("SET DB error: %v\n", err)
			return err
		}
	}

	return nil
}

func NotificationWithinBounds(s, w, n, e float64) ([]GeoJsonFeature, error) {
	// Connect Tile38
	c := pool.Get()
	defer c.Close()

	currentTime := time.Now().Unix()
	now := fmt.Sprintf("%v", currentTime)
	before60Sec := fmt.Sprintf("%v", currentTime-60)
	ret, err := db_withinBounds(c, "notification", s, w, n, e, "WHERE", "time", before60Sec, now)
	if err != nil {
		log.Fatalf("DB WITHIN error: %v\n", err)
		return nil, err
	}
	//fmt.Printf("%s\n", ret)

	return ret, nil
}

func DropNotification() error {
	// Connect Tile38
	c := pool.Get()
	defer c.Close()

	err := db_drop(c, "notification")
	if err != nil {
		log.Fatalf("DB DROP error: %v\n", err)
		return err
	}
	return nil
}
