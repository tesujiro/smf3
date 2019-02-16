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

func TestAPIFlyers(t *testing.T) {
	// No log
	log.SetOutput(ioutil.Discard)

	now := time.Now().Unix()
	flyers := []*db.Flyer{
		&db.Flyer{OwnerID: 1, Title: "title01", ValidPeriod: 3600, StartAt: now, EndAt: now + 3600, Lat: 0, Lon: 0, Distance: 100, Stocked: 100, Delivered: 0},
		&db.Flyer{OwnerID: 1, Title: "title01", ValidPeriod: 3600, StartAt: now, EndAt: now + 3600, Lat: 0, Lon: 0, Distance: 100, Stocked: 100, Delivered: 0},
		&db.Flyer{OwnerID: 1, Title: "title01", ValidPeriod: 3600, StartAt: now, EndAt: now + 3600, Lat: 0, Lon: 0, Distance: 100, Stocked: 100, Delivered: 0},
	}
	data := []*db.Flyer{flyers[0], flyers[1]}
	tests := []struct {
		method         string
		url            string
		bounds         map[string]string
		flyer          *db.Flyer
		body           string
		status         int
		header         map[string]string
		expectedLength int
	}{
		{method: "GET", url: "/api/flyers", bounds: map[string]string{"south": "0", "west": "0", "north": "0", "east": "0"}, expectedLength: 2},
		{method: "GET", url: "/api/flyers", bounds: map[string]string{}, status: 500},
		{method: "GET", url: "/api/flyers", bounds: map[string]string{"xxx": "0"}, status: 500},
		{method: "GET", url: "/api/flyers", bounds: map[string]string{"south": "0", "west": "0", "north": "0", "east": "xxx"}, status: 500},
		{method: "GET", url: "/api/flyers", bounds: map[string]string{"south": "0", "west": "0", "north": "0"}, status: 500},
		{method: "GET", url: "/api/flyers", bounds: map[string]string{"south": "1", "SOUTH": "1", "west": "0", "north": "1", "EAST": "1"}, status: 500},
		{method: "POST", url: "/api/flyers", flyer: flyers[2], expectedLength: 3},
	}

	// SET DATA
	for _, flyer := range data {
		flyer.ID = db.NewFlyerID()
		if err := flyer.Set(); err != nil {
			t.Fatalf("Set Flyer error: (%v) flyer:%v\n", err, flyer)
		}
		if err := flyer.Sethook("http://dummy/dummy"); err != nil {
			t.Fatalf("Set Flyer error: (%v) flyer:%v\n", err, flyer)
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
			if reqBody, err = json.Marshal(test.flyer); err != nil {
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
			actualFlyers, err := db.ScanValidFlyers(now)
			if err != nil {
				t.Errorf("Test[%v] ScanValidFLyers error: %v", test_number, err)
			}
			if len(actualFlyers) != test.expectedLength {
				t.Errorf("Test[%v] check flyers length error. expected: %v actual: %v", test_number, test.expectedLength, len(actualFlyers))
			}

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
			var actualFlyers []*db.Flyer
			if err := json.Unmarshal(data, &actualFlyers); err != nil {
				t.Errorf("Test[%v] Response body json.Unmarshal error: %v", test_number, err)
			}
			if len(actualFlyers) != test.expectedLength {
				t.Errorf("Test[%v] response length error. expected: %v actual: %v", test_number, test.expectedLength, len(actualFlyers))
			}
		}
	}

	//DELETE TEST DATA
	db.DropFlyer()
	log.SetOutput(os.Stdout)
}
