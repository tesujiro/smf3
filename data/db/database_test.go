package db

import (
	"encoding/json"
	"reflect"
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

	features := []*GeoJsonFeature{
		&GeoJsonFeature{ // [0]
			Type: "Feature",
			Geometry: &Geometry{
				Type:        "Point",
				Coordinates: [2]float64{1.2345, 100.2003},
			},
			Properties: map[string]interface{}{
				"id": "ID",
			},
		},
		&GeoJsonFeature{ // [1]
			Type: "Feature",
			Geometry: &Geometry{
				Type:        "Point",
				Coordinates: [2]float64{1.2345, 100.2003},
			},
			Properties: map[string]interface{}{
				"id":   "ID",
				"some": "something",
			},
		},
	}

	tests := []struct {
		key             string
		id              string
		feature         *GeoJsonFeature
		expectedFeature *GeoJsonFeature
		jset_path       string
		jset_value      interface{}
		args            []interface{}
	}{
		{key: "test-key-1", id: "test-id-1", feature: features[0], args: []interface{}{}},
		{key: "test-key-1", id: "test-id-1", feature: features[0], args: []interface{}{"FIELD", "f1", 10}},
		{key: "test-key-1", id: "test-id-1", feature: features[0], args: []interface{}{"FIELD", "f1", 10, "FIELD", "f2", 20.345}},
		{key: "test-key-1", id: "test-id-1", feature: features[0], args: []interface{}{"FIELD", "f1", 10, "NX"}},
		{key: "test-key-1", id: "test-id-1", feature: features[0], expectedFeature: features[1], jset_path: "properties.some", jset_value: "something", args: []interface{}{}},
	}

	c := pool.Get()
	defer c.Close()
	for i, test := range tests {
		var test_json string
		b, err := json.Marshal(test.feature)
		if err != nil {
			t.Fatalf("JSON Marshal error: %v", err)
		}
		test_json = string(b)
		if test.expectedFeature == nil {
			test.expectedFeature = test.feature
		}
		//fmt.Printf("test_json:%v\n", test_json)
		if err := db_set_json(c, test.key, test.id, test_json, test.args...); err != nil {
			t.Fatalf("SET DB error: %v", err)
		}
		if test.jset_path != "" {
			if err := db_jset(c, test.key, test.id, test.jset_path, test.jset_value, test.args...); err != nil {
				t.Fatalf("JSET DB error: %v", err)
			}
		}
		if b, err := db_get(c, test.key, test.id); err != nil {
			t.Fatalf("GET DB error: %v\n", err)
		} else {
			var actual *GeoJsonFeature
			err := json.Unmarshal(b, &actual)
			if err != nil {
				t.Fatalf("JSON Unmarshal error: %v", err)
			}
			if !reflect.DeepEqual(test.expectedFeature, actual) {
				t.Errorf("Case:[%v] received: %v - expected: %v", i, actual, test.expectedFeature)
			}
		}
		//t.Logf("Case:[%v] received: %v - expected: %v", i, string(b), expected_json)
		//defer db_del(c, test.key, test.id)
		if err := db_del(c, test.key, test.id); err != nil {
			t.Fatalf("DELETE DB error: %v\n", err)
		}
	}
}

func TestScan(t *testing.T) {
	data := []struct {
		key      string
		id       string
		lat, lon float64
		time     float64
	}{
		{key: "test-key1", id: "id1", lat: 0, lon: 0, time: 100},
		{key: "test-key1", id: "id2", lat: 1.23, lon: 4.56, time: 10},
		{key: "test-key1", id: "id3", lat: 1, lon: 2, time: 500},
		{key: "test-key1", id: "id4", lat: 0, lon: 0, time: 0},
	}
	tests := []struct {
		key         string
		args        []interface{}
		expectedLen int
	}{
		{key: "test-key1", expectedLen: 4},
		{key: "test-key2", expectedLen: 0},
		{key: "test-key1", args: []interface{}{"WHERE", "time", 0, 100}, expectedLen: 3},
		{key: "test-key1", args: []interface{}{"WHERE", "time", 300, 500}, expectedLen: 1},
		{key: "test-key2", args: []interface{}{"WHERE", "time", 0, 100}, expectedLen: 0},
	}

	c := pool.Get()
	defer c.Close()
	for _, record := range data {
		feature := &GeoJsonFeature{ // [0]
			Type: "Feature",
			Geometry: &Geometry{
				Type:        "Point",
				Coordinates: [2]float64{record.lon, record.lat},
			},
			Properties: map[string]interface{}{
				"id":   record.id,
				"time": record.time,
			},
		}
		var test_json string
		b, err := json.Marshal(feature)
		if err != nil {
			t.Fatalf("JSON Marshal error: %v", err)
		}
		test_json = string(b)
		if err := db_set_json(c, record.key, record.id, test_json, "FIELD", "time", record.time); err != nil {
			t.Fatalf("SET DB error: %v", err)
		}
	}
	for i, test := range tests {
		features, err := db_scan(c, test.key, test.args...)
		if err != nil {
			t.Fatalf("SCAN DB error: %v", err)
		}
		if len(features) != test.expectedLen {
			t.Errorf("Case:[%v] received: %v - expected: %v", i, len(features), test.expectedLen)
		}
	}
	for _, key := range []string{"test-key1"} {
		err := db_drop(c, key)
		if err != nil {
			t.Fatalf("DROP DB (key:%v) error: %v", key, err)
		}
	}
}

func TestWithin(t *testing.T) {
	data := []struct {
		key      string
		id       string
		lat, lon float64
		time     float64
	}{
		{key: "test-key1", id: "id1", lat: 0, lon: 0, time: 100},
		{key: "test-key1", id: "id2", lat: 1.23, lon: 4.56, time: 10},
		{key: "test-key1", id: "id3", lat: 1, lon: 2, time: 500},
		{key: "test-key1", id: "id4", lat: 0, lon: 0, time: 0},
	}
	tests := []struct {
		key              string
		area             string
		s, w, n, e       float64
		lat, lon, meters float64
		args             []interface{}
		expectedLen      int
	}{
		{key: "test-key1", area: "bounds", s: 0, w: 0, n: 2, e: 2, expectedLen: 3},
		{key: "test-key2", area: "bounds", s: 0, w: 0, n: 2, e: 2, expectedLen: 0},
		{key: "test-key1", area: "bounds", s: 0, w: 0, n: 2, e: 2, args: []interface{}{"WHERE", "time", 0, 100}, expectedLen: 2},
		{key: "test-key1", area: "circle", lat: 0, lon: 0, meters: 500000, expectedLen: 3},
		{key: "test-key2", area: "circle", lat: 0, lon: 0, meters: 500000, expectedLen: 0},
		{key: "test-key1", area: "circle", lat: 0, lon: 0, meters: 500000, args: []interface{}{"WHERE", "time", 0, 100}, expectedLen: 2},
	}

	c := pool.Get()
	defer c.Close()
	for _, record := range data {
		feature := &GeoJsonFeature{ // [0]
			Type: "Feature",
			Geometry: &Geometry{
				Type:        "Point",
				Coordinates: [2]float64{record.lon, record.lat},
			},
			Properties: map[string]interface{}{
				"id":   record.id,
				"time": record.time,
			},
		}
		var test_json string
		b, err := json.Marshal(feature)
		if err != nil {
			t.Fatalf("JSON Marshal error: %v", err)
		}
		test_json = string(b)
		if err := db_set_json(c, record.key, record.id, test_json, "FIELD", "time", record.time); err != nil {
			t.Fatalf("SET DB error: %v", err)
		}
	}
	for i, test := range tests {
		switch test.area {
		case "bounds":
			features, err := db_withinBounds(c, test.key, test.s, test.w, test.n, test.e, test.args...)
			if err != nil {
				t.Fatalf("WITHIN DB error: %v", err)
			}
			if len(features) != test.expectedLen {
				t.Errorf("Case:[%v] received: %v - expected: %v", i, len(features), test.expectedLen)
			}
		case "circle":
			features, err := db_withinCircle(c, test.key, test.lat, test.lon, test.meters, test.args...)
			if err != nil {
				t.Fatalf("WITHIN DB error: %v", err)
			}
			if len(features) != test.expectedLen {
				t.Errorf("Case:[%v] received: %v - expected: %v", i, len(features), test.expectedLen)
			}
		}
	}
	for _, key := range []string{"test-key1"} {
		err := db_drop(c, key)
		if err != nil {
			t.Fatalf("DROP DB (key:%v) error: %v", key, err)
		}
	}
}
