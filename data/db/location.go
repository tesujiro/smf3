package db

import (
	"encoding/json"
	"fmt"
	"log"
)

type Location struct {
	ID   int64
	Lat  float64
	Lon  float64
	Time string
}

func (loc *Location) geoJson() (string, error) {
	feature := &GeoJsonFeature{
		Type: "Feature",
		Geometry: &Geometry{
			Type:        "Point",
			Coordinates: [2]float64{loc.Lon, loc.Lat},
		},
		Properties: map[string]interface{}{
			"id":   loc.ID,
			"time": loc.Time,
		},
	}

	json, err := json.Marshal(feature)
	if err != nil {
		log.Printf("Notification feature marshal error: %v\n", err)
		return "", err
	}
	return string(json), nil
}

func (loc *Location) Set() error {
	// Connect Tile38
	/*
		c, err := db_connect()
		if err != nil {
			log.Fatalf("Connect tile38-server\n")
			return err
		}
	*/
	c := pool.Get()
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

func LocationWithinBounds(s, w, n, e float64) ([]GeoJsonFeature, error) {
	// Connect Tile38
	c := pool.Get()
	defer c.Close()

	ret, err := db_withinBounds(c, "location", s, w, n, e)
	if err != nil {
		log.Fatalf("DB WITHIN error: %v\n", err)
		return nil, err
	}
	//fmt.Printf("%s\n", ret)

	return ret, nil
}

func LocationWithinCircle(lat, lon, meter float64, args ...interface{}) ([]GeoJsonFeature, error) {
	// Connect Tile38
	c := pool.Get()
	defer c.Close()

	ret, err := db_withinCircle(c, "location", lat, lon, meter, args...)
	if err != nil {
		log.Fatalf("DB WITHIN error: %v\n", err)
		return nil, err
	}
	//fmt.Printf("%s\n", ret)

	return ret, nil
}

func DropLocation() error {
	// Connect Tile38
	c := pool.Get()
	defer c.Close()

	err := db_drop(c, "location")
	if err != nil {
		log.Fatalf("DB DROP error: %v\n", err)
		return err
	}
	return nil
}
