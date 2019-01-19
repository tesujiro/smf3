package db

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestConnect(t *testing.T) {
	conn, err := db_connect()
	if err != nil {
		t.Fatalf("failed to connect tile38-server:%v\n", err)
	}
	defer conn.Close()
}

func TestSetJson(t *testing.T) {

	features := []GeoJsonFeature{
		GeoJsonFeature{
			Type: "Feature",
			Geometry: &Geometry{
				Type:        "Point",
				Coordinates: [2]float64{1.2345, 100.2003},
			},
			Properties: map[string]interface{}{
				"id": "ID",
			},
		},
	}

	tests := []struct {
		key     string
		id      string
		feature GeoJsonFeature
		args    []interface{}
	}{
		{key: "test-key-1", id: "test-id-1", feature: features[0], args: []interface{}{}},
	}

	c := pool.Get()
	defer c.Close()
	for i, test := range tests {
		b, err := json.Marshal(test.feature)
		if err != nil {
			t.Fatalf("JSON Marshal error: %v", err)
		}
		json := string(b)
		fmt.Printf("json:%v\n", json)
		if err := db_set_json(c, test.key, test.id, json, test.args...); err != nil {
			t.Fatalf("SET DB error: %v", err)
		}
		if b, err := db_get(c, test.key, test.id); err != nil {
			t.Fatalf("GET DB error: %v\n", err)
		} else if string(b) != json {
			t.Errorf("Case:[%v] received: %v - expected: %v", i, string(b), json)
		}
		defer db_del(c, test.key, test.id)
	}
}
