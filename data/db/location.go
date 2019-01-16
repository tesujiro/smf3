package db

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"
)

type Location struct {
	ID   int64
	Lat  float64
	Lon  float64
	Time string
}

func (loc *Location) geoJson() (string, error) {
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
		"id": {{.ID}},
		"time": "{{.Time}}"
	}
}`
	t := template.Must(template.New("location").Parse(json_template))
	t.Funcs(template.FuncMap{
		"now":     func() string { return time.Now().String() },
		"toupper": strings.ToUpper,
	})

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, loc); err != nil {
		return "", err
	}
	return tpl.String(), nil
}

func (loc *Location) Set() error {
	// Connect Tile38
	c, err := db_connect()
	if err != nil {
		log.Fatalf("Connect tile38-server\n")
		return err
	}
	defer c.Close()

	if json, err := loc.geoJson(); err != nil {
		fmt.Printf("location.geoJson() error: %v\n", err)
		return err
	} else {
		//fmt.Printf("GeoJSON:%v\n", json)
		err = db_set_json(c, "location", fmt.Sprintf("%v", loc.ID), json)
		if err != nil {
			log.Fatalf("SET DB error: %v\n", err)
			return err
		}
	}

	return nil
}

func WithinLocation(s, w, n, e float64) ([]interface{}, error) {
	// Connect Tile38
	c, err := db_connect()
	if err != nil {
		log.Fatalf("Connect tile38-server\n")
		return nil, err
	}
	defer c.Close()

	ret, err := db_within(c, "location", s, w, n, e)
	if err != nil {
		log.Fatalf("DB WITHIN error: %v\n", err)
		return nil, err
	}
	//fmt.Printf("%s\n", ret)

	return ret, nil
}

func DropLocation() error {
	// Connect Tile38
	c, err := db_connect()
	if err != nil {
		log.Fatalf("Connect tile38-server\n")
		return err
	}
	defer c.Close()

	err = db_drop(c, "location")
	if err != nil {
		log.Fatalf("DB DROP error: %v\n", err)
		return err
	}
	return nil
}
