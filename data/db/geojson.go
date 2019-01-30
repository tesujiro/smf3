package db

import (
	"encoding/json"
	"fmt"
)

/*
type ObjectType int
const (
	GeometryObject = iota
	FeatureObject
	FeatureCollectionObject
)

type GeoJsonType int
const (
	Point = iota
	MultiPoint
	LineString
	MultiLineString
	Polygon
	MultiPolygon
	GeometryCollection
	Feature
	FeatureCollection
)
*/

type GeoJsonFeature struct {
	Type       string                 `json:"type,omitempty"`
	Geometry   *Geometry              `json:"geometry,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

type Geometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates,omitempty"`
	//Geometry    *Geometry       `json:"geometries,omitempty"`
}

// CoordinatesObject
type Point [2]float64

type LineString []Point

type Polygon []LineString

func (g *Geometry) GetCoordinatesObject() (interface{}, error) {
	var object interface{}
	switch g.Type {
	case "Point":
		object = new(Point)
	case "LineString":
		object = new(LineString)
	case "Polygon":
		object = new(Polygon)
	default:
		return nil, fmt.Errorf("Unknown type: %v", g.Type)
	}
	err := json.Unmarshal(g.Coordinates, &object)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal error:%v coordinates:%s", err, g.Coordinates)
	}
	//fmt.Printf("object:%v\n", object)
	return object, nil
}
