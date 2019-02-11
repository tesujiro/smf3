package db

import (
	"encoding/json"
	"fmt"
	"log"
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
	Stocked     int     `json:"stocked"`
	Delivered   int     `json:"delivered"`
}

var currentFlyerID int64 = 0

func NewFlyerID() int64 {
	currentFlyerID++
	return currentFlyerID
}

func (fly *Flyer) geoJson() (string, error) {
	feature := &GeoJsonFeature{
		Type: "Feature",
		Geometry: &Geometry{
			Type:        "Point",
			Coordinates: []byte(fmt.Sprintf("[%v,%v]", fly.Lon, fly.Lat)),
		},
		Properties: map[string]interface{}{
			"id":          fly.ID,
			"ownerId":     fly.OwnerID,
			"title":       fly.Title,
			"validPeriod": fly.ValidPeriod,
			"startAt":     fly.StartAt,
			"endAt":       fly.EndAt,
			"distance":    fly.Distance,
			"stocked":     fly.Stocked,
			"delivered":   fly.Delivered,
		},
	}

	json, err := json.Marshal(feature)
	if err != nil {
		log.Printf("Notification feature marshal error: %v\n", err)
		return "", err
	}
	return string(json), nil
}

func geoJson2Flyer(feature GeoJsonFeature) (*Flyer, error) {
	pj, err := json.Marshal(feature.Properties)
	if err != nil {
		log.Fatalf("Marshal feature error: %v\n", err)
		return nil, err
	}
	//fmt.Printf("property json:%s\n", pj)

	var f Flyer
	err = json.Unmarshal(pj, &f)
	if err != nil {
		log.Fatalf("Unmarshal feature error: %v\n", err)
		return nil, err
	}
	c, err := feature.Geometry.GetCoordinatesObject()
	if err != nil {
		log.Fatalf("Unmarshal coordinates error: %v\n", err)
		return nil, err
	}
	point, ok := c.(*Point)
	if !ok {
		log.Fatalf("Coordinates conversion error: not point  c=%#v\n", c)
		return nil, err
	}
	f.Lat = point[1]
	f.Lon = point[0]

	return &f, nil
}

func GetFlyer(id string) (*Flyer, error) {
	c := pool.Get()
	defer c.Close()

	//id_num, _ := strconv.ParseInt(id, 10, 64)
	//b, err := db_get(c, "flyer", id_num)
	b, err := db_get(c, "flyer", id)
	if err != nil {
		log.Fatalf("GET DB error: %v\n", err)
		return nil, err
	}
	if b == nil {
		log.Printf("GET flyer not found\n")
		return nil, nil
	}

	var feature GeoJsonFeature
	err = json.Unmarshal(b, &feature)
	if err != nil {
		log.Fatalf("Unmarshal error: %v\n", err)
		return nil, err
	}
	f, err := geoJson2Flyer(feature)
	if err != nil {
		log.Fatalf("geoJson2Flyer error: %v\n", err)
		return nil, err
	}

	return f, nil
}

func (fly *Flyer) Set() error {
	// Connect Tile38
	c := pool.Get()
	defer c.Close()

	if json, err := fly.geoJson(); err != nil {
		log.Printf("flyer.geoJson() error: %v\n", err)
		return err
	} else {
		//fmt.Printf("GeoJSON:%v\n", json)
		err = db_set_json(c, "flyer", fmt.Sprintf("%v", fly.ID), json, "FIELD", "start", fly.StartAt, "FIELD", "end", fly.EndAt, "EX", fly.ValidPeriod)
		if err != nil {
			log.Fatalf("SET DB error: %v\n", err)
			return err
		}
	}

	return nil
}

func (fly *Flyer) Jset(path string, value interface{}) error {
	// Connect Tile38
	c := pool.Get()
	defer c.Close()

	//fmt.Printf("GeoJSON:%v\n", json)
	err := db_jset(c, "flyer", fmt.Sprintf("%v", fly.ID), path, value)
	if err != nil {
		log.Fatalf("JSET DB error: %v\n", err)
		return err
	}

	return nil
}

func (fly *Flyer) Sethook(endpoint string) error {
	c := pool.Get()
	defer c.Close()

	//fmt.Printf("GeoJSON:%v\n", json)
	hookname := fmt.Sprintf("flyerhook:%v", fly.ID)
	err := db_sethook(c, hookname, endpoint, "EX", fly.ValidPeriod, "NEARBY", "location", "FENCE", "DETECT", "enter", "POINT", fly.Lat, fly.Lon, fly.Distance)
	if err != nil {
		log.Fatalf("SETHOOK DB error: %v\n", err)
		return err
	}

	return nil
}

func ScanValidFlyers(currentTime int64) ([]Flyer, error) {
	// Connect Tile38
	c := pool.Get()
	defer c.Close()

	time := fmt.Sprintf("%v", currentTime)
	ret, err := db_scan(c, "flyer", "WHERE", "start", "-inf", time, "WHERE", "end", time, "+inf")
	if err != nil {
		log.Fatalf("DB Scan error: %v\n", err)
		return nil, err
	}
	//fmt.Printf("%s\n", ret)

	flyers := make([]Flyer, len(ret))
	for i, feature := range ret {
		//fmt.Printf("feature:%#v\n", feature)

		// TODO: call geoJson2Flyer(feature GeoJsonFeature)
		pj, err := json.Marshal(feature.Properties)
		if err != nil {
			log.Fatalf("Marshal feature error: %v\n", err)
			return nil, err
		}
		//fmt.Printf("property json:%s\n", pj)

		var f Flyer
		err = json.Unmarshal(pj, &f)
		if err != nil {
			log.Fatalf("Unmarshal feature error: %v\n", err)
			return nil, err
		}
		c, err := feature.Geometry.GetCoordinatesObject()
		if err != nil {
			log.Fatalf("Unmarshal coordinates error: %v\n", err)
			return nil, err
		}
		point, ok := c.(*Point)
		if !ok {
			log.Fatalf("Coordinates conversion error: not point  c=%#v\n", c)
			return nil, err
		}
		f.Lat = point[1]
		f.Lon = point[0]
		//fmt.Printf("peoperty :%#v\n", f)

		flyers[i] = f
	}

	return flyers, nil
}

func FlyerWithinBounds(s, w, n, e float64, args ...interface{}) ([]GeoJsonFeature, error) {
	// Connect Tile38
	c := pool.Get()
	defer c.Close()

	ret, err := db_withinBounds(c, "flyer", s, w, n, e, args...)
	if err != nil {
		log.Fatalf("DB WITHIN error: %v\n", err)
		return nil, err
	}
	//fmt.Printf("%s\n", ret)

	return ret, nil
}

func DropFlyer() error {
	// Connect Tile38
	c := pool.Get()
	defer c.Close()

	err := db_drop(c, "flyer")
	if err != nil {
		log.Fatalf("DB DROP error: %v\n", err)
		return err
	}
	err = db_pdelhook(c, "flyerhook:*")
	if err != nil {
		log.Fatalf("DB PDELHOOK error: %v\n", err)
		return err
	}
	return nil
}
