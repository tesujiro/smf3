package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
)

type Notification struct {
	ID           int64
	FlyerID      int64
	UserID       int64
	Lat          float64
	Lon          float64
	DeliveryTime int64
}

func (n *Notification) geoJson() (string, error) {
	json_template := `{
	"type": "Feature",
	"geometry": {
		"type": "Point",
		"coordinates": [
			{{.Lon}},
			{{.Lat}}
		]
	},
	"properties": {
		"id":      {{.ID}},
		"flyerId": {{.FlyerID}},
		"userId":  {{.UserID}},
		"deliveryTime": {{.DeliveryTime}}
	}
}`
	t := template.Must(template.New("notification").Parse(json_template))

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, n); err != nil {
		return "", err
	}
	return tpl.String(), nil
}

func GetNotification(id string) (*Notification, error) {
	// Connect Tile38
	c, err := db_connect()
	if err != nil {
		log.Fatalf("Connect tile38-server\n")
		return nil, err
	}
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
func (n *Notification) Set() error {
	// Connect Tile38
	c, err := db_connect()
	if err != nil {
		log.Fatalf("Connect tile38-server\n")
		return err
	}
	defer c.Close()

	if json, err := n.geoJson(); err != nil {
		fmt.Printf("notification.geoJson() error: %v\n", err)
		return err
	} else {
		//fmt.Printf("GeoJSON:%v\n", json)
		err = db_set_json(c, "notification", fmt.Sprintf("%v", n.ID), json, "NX") // NX: IF NOT EXISTS
		if err != nil {
			log.Fatalf("SET DB error: %v\n", err)
			return err
		}
	}

	return nil
}

func NotificationWithinBounds(s, w, n, e float64) ([]interface{}, error) {
	// Connect Tile38
	c, err := db_connect()
	if err != nil {
		log.Fatalf("Connect tile38-server\n")
		return nil, err
	}
	defer c.Close()

	ret, err := db_withinBounds(c, "notification", s, w, n, e)
	if err != nil {
		log.Fatalf("DB WITHIN error: %v\n", err)
		return nil, err
	}
	//fmt.Printf("%s\n", ret)

	return ret, nil
}

func DropNotification() error {
	// Connect Tile38
	c, err := db_connect()
	if err != nil {
		log.Fatalf("Connect tile38-server\n")
		return err
	}
	defer c.Close()

	err = db_drop(c, "notification")
	if err != nil {
		log.Fatalf("DB DROP error: %v\n", err)
		return err
	}
	return nil
}
