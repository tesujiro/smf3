package db

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Notification struct {
	ID           string // flyerID+":"+userID
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

/*
var currentNotificationID int64 = 0

func NewNotificationID() int64 {
	currentNotificationID++
	return currentNotificationID
}
*/

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

func ExistNotification(id string) (bool, error) {
	// Connect Tile38
	c := pool.Get()
	defer c.Close()

	b, err := db_get(c, "notification", id)
	if err != nil {
		log.Fatalf("GET DB error: %v\n", err)
		return false, err
	}
	if b == nil {
		return false, nil
	}
	return true, nil

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

func NotificationFeaturesWithinBounds(s, w, n, e float64) ([]GeoJsonFeature, error) {
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
	fmt.Printf("NotificationFeaturesWithinBounds: %v\n", len(ret))

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
