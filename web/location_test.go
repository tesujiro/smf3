package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/tesujiro/smf3/data/db"
)

func TestAPILocations(t *testing.T) {
	// No log
	log.SetOutput(ioutil.Discard)

	now := time.Now().Unix()
	locations := []*db.Location{
		&db.Location{Lat: 0, Lon: 0, Time: now},
		&db.Location{Lat: 1, Lon: 1, Time: now},
		&db.Location{Lat: 2, Lon: 2, Time: now},
		&db.Location{Lat: 3, Lon: 3, Time: now},
	}
	data := []*db.Location{locations[0], locations[1], locations[2]}
	tests := []struct {
		method         string
		url            string
		bounds         map[string]string
		location       *db.Location
		body           string
		status         int
		header         map[string]string
		expectedLength int
	}{
		{method: "GET", url: "/api/locations", bounds: map[string]string{"south": "0", "west": "0", "north": "1", "EAST": "1"}, expectedLength: 2},
		{method: "GET", url: "/api/locations", bounds: map[string]string{}, status: 500},
		{method: "GET", url: "/api/locations", bounds: map[string]string{"xxx": "0"}, status: 500},
		{method: "GET", url: "/api/locations", bounds: map[string]string{"south": "0", "west": "0", "north": "0", "east": "xxx"}, status: 500},
		{method: "GET", url: "/api/locations", bounds: map[string]string{"south": "0", "west": "0", "north": "0"}, status: 500},
		{method: "GET", url: "/api/locations", bounds: map[string]string{"south": "1", "SOUTH": "1", "west": "0", "north": "1", "EAST": "1"}, status: 500},
		{method: "POST", url: "/api/locations", location: locations[2], expectedLength: 3},
	}

	// SET DATA
	for _, location := range data {
		location.ID = db.NewLocationID()
		if err := location.Set(); err != nil {
			t.Fatalf("Set Location error: (%v) location:%v\n", err, location)
		}
	}

	// TEST
	srv := newServer()
	srv.routes()
	for test_number, test := range tests {
		if test.method == http.MethodGet {
			test.url = fmt.Sprintf("%v?", test.url)
			for k, v := range test.bounds {
				test.url = fmt.Sprintf("%v&%v=%v", test.url, k, v)
			}
		}
		var reqBody []byte
		if test.method == http.MethodPost {
			var err error
			if reqBody, err = json.Marshal(test.location); err != nil {
				t.Errorf("Test[%v] failed to marshal request body: %v", test_number, err)
			}
		}
		req, err := http.NewRequest(test.method, test.url, bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Length", fmt.Sprintf("%v", len(reqBody)))
		if err != nil {
			t.Errorf("Test[%v] failed http.NewRequest %v", test_number, err)
		}
		w := httptest.NewRecorder()
		srv.router.ServeHTTP(w, req)
		r := w.Result()
		//fmt.Printf("Result:%#v\n", r)
		if test.status == 0 && r.StatusCode != http.StatusOK ||
			test.status != 0 && r.StatusCode != test.status {
			fmt.Printf("result:%#v\n", r)
			t.Errorf("Test[%v] method:%v url:%v StatusCode:%v", test_number, test.method, test.url, r.StatusCode)
		}
		if test.status != 0 {
			continue
		}
		//fmt.Printf("header.Location:%#v\n", r.Header["Location"])
		switch test.method {
		case http.MethodPost:
			// Check Database
			/*
				actualLocations, err := db.ScanLocations(now)
				if err != nil {
					t.Errorf("Test[%v] ScanValidFLyers error: %v", test_number, err)
				}
				if len(actualLocations) != test.expectedLength {
					t.Errorf("Test[%v] check location length error. expected: %v actual: %v", test_number, test.expectedLength, len(actualLocations))
				}
			*/

		case http.MethodGet:
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fmt.Printf("result:%#v\n", r)
				t.Errorf("Test[%v] method:%v url:%v Error by ioutil.ReadAll(). %v", test_number, test.method, test.url, err)
			}
			if test.body != "" && string(data) != test.body {
				fmt.Printf("result:%#v\n", r)
				t.Errorf("Test[%v] method:%v url:%v Data Error. [%v]", test_number, test.method, test.url, string(data))
			}
			//fmt.Printf("Body:%v\n", string(data))
			var actualLocations []*db.Location
			if err := json.Unmarshal(data, &actualLocations); err != nil {
				t.Errorf("Test[%v] Response body json.Unmarshal error: %v", test_number, err)
			}
			if len(actualLocations) != test.expectedLength {
				t.Errorf("Test[%v] response length error. expected: %v actual: %v", test_number, test.expectedLength, len(actualLocations))
			}
		}
	}

	//DELETE TEST DATA
	db.DropLocation()
	log.SetOutput(os.Stdout)
}
