package db

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"
)

type Flyer struct {
	ID          int64   `json:"Id"`
	OwnerID     int64   `json:"storeId"`
	Title       string  `json:"title"`
	ValidPeriod int64   `json:"validPeriod"`
	StartAt     int64   `json:"startAt"`
	EndAt       int64   `json:"endAt"`
	Lat         float64 `json:"latitude"`
	Lon         float64 `json:"longitude"`
	Distance    float64 `json:"distance"`
	Stocked     int     `json:"pieces"`
	Delivered   int
}

func (fly *Flyer) geoJson() (string, error) {
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
		"id":         {{.ID}},
		"ownerId":    {{.OwnerID}},
		"title":     "{{.Title}}",
		"validPeriod": {{.ValidPeriod}},
		"startAt":    {{.StartAt}},
		"endAt":      {{.EndAt}},
		"distance":   {{.Distance}},
		"stocked":    {{.Stocked}},
		"delivered":  {{.Delivered}}
	}
}`
	t := template.Must(template.New("flyer").Parse(json_template))
	t.Funcs(template.FuncMap{
		"now":     func() string { return time.Now().String() },
		"toupper": strings.ToUpper,
	})

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, fly); err != nil {
		return "", err
	}
	return tpl.String(), nil
}

func (fly *Flyer) Set() error {
	// Connect Tile38
	c, err := db_connect()
	if err != nil {
		log.Fatalf("Connect tile38-server\n")
		return err
	}
	defer c.Close()

	if json, err := fly.geoJson(); err != nil {
		fmt.Printf("flyer.geoJson() error: %v\n", err)
		return err
	} else {
		//fmt.Printf("GeoJSON:%v\n", json)
		err = db_set_json(c, "flyer", fmt.Sprintf("%v", fly.ID), json, "FIELD", "start", fly.StartAt, "FIELD", "end", fly.EndAt)
		if err != nil {
			log.Fatalf("SET DB error: %v\n", err)
			return err
		}
	}

	return nil
}

func ScanValidFlyers(currentTime int64) (string, error) {
	// Connect Tile38
	c, err := db_connect()
	if err != nil {
		log.Fatalf("Connect tile38-server\n")
		return "", err
	}
	defer c.Close()

	time := fmt.Sprintf("%v", currentTime)
	ret, err := db_scan(c, "flyer", "WHERE", "start", "-inf", time, "WHERE", "end", time, "+inf")
	if err != nil {
		log.Fatalf("DB Scan error: %v\n", err)
		return "", err
	}
	//fmt.Printf("%s\n", ret)

	return ret, nil
}

func WithinFlyer(s, w, n, e float64) (string, error) {
	// Connect Tile38
	c, err := db_connect()
	if err != nil {
		log.Fatalf("Connect tile38-server\n")
		return "", err
	}
	defer c.Close()

	ret, err := db_within(c, "flyer", s, w, n, e)
	if err != nil {
		log.Fatalf("DB WITHIN error: %v\n", err)
		return "", err
	}
	//fmt.Printf("%s\n", ret)

	return ret, nil
}

func DropFlyer() error {
	// Connect Tile38
	c, err := db_connect()
	if err != nil {
		log.Fatalf("Connect tile38-server\n")
		return err
	}
	defer c.Close()

	err = db_drop(c, "flyer")
	if err != nil {
		log.Fatalf("DB DROP error: %v\n", err)
		return err
	}
	return nil
}
