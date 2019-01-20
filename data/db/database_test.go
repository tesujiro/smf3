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
