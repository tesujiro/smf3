package db

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

type Location struct {
	ID   int64
	Lat  float64
	Lon  float64
	Time string
}

func db_connect() (redis.Conn, error) {
	c, err := redis.Dial("tcp", ":9851")
	if err != nil {
		log.Fatalf("Could not connect: %v\n", err)
		return nil, err
	}
	return c, nil
}

func db_set_json(c redis.Conn, key, id, json string) error {
	_, err := c.Do("SET", key, id, "OBJECT", json)
	//fmt.Printf("%s\n", ret)
	return err
}

func db_get(c redis.Conn, key, id string) (string, error) {
	ret, err := c.Do("GET", key, id)
	if err != nil {
		return "", err
	}
	return string(ret.([]byte)), err
}

func db_scan(c redis.Conn, key string) ([]string, error) {
	ret, err := c.Do("SCAN", key)
	if err != nil {
		return []string{}, err
	}

	records := ret.([]interface{})[1].([]interface{})
	jsons := make([]string, len(records))
	for i, b := range records {
		str := string(b.([]interface{})[1].([]byte))
		jsons[i] = str
		//fmt.Printf("%v:%s\n", i, str)
	}

	/*
		values, err := c.Do("SCAN", key)
		values, err := redis.Values(c.Do("SCAN", key))
		if err != nil {
			return "", err
		}

		fmt.Printf("len(values):%v\n", len(values))
		//for i, val := range values {
		for i, val := range values[1].([]interface{}) {
			fmt.Printf("[%v]: %#v\n", i, val)
		}
	*/

	return jsons, err
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
		"ID": {{.ID}},
		"Time": "{{.Time}}"
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
		log.Fatalf("Start tile38-server\n")
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
func ScanLocation() ([]string, error) {
	// Connect Tile38
	c, err := db_connect()
	if err != nil {
		log.Fatalf("Start tile38-server\n")
		return []string{}, err
	}
	defer c.Close()

	ret, err := db_scan(c, "location")
	if err != nil {
		log.Fatalf("DB SCAN error: %v\n", err)
		return []string{}, err
	}
	//fmt.Printf("%s\n", ret)

	return ret, nil
}
