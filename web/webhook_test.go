package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/tesujiro/smf3/data/db"
)

func TestE2EWebhook(t *testing.T) {
	// No log
	//log.SetOutput(ioutil.Discard)
	log.SetFlags(log.Lmicroseconds)

	now := time.Now().Unix()
	flyers := []*db.Flyer{
		&db.Flyer{OwnerID: 1, Title: "title01", ValidPeriod: 3600, Lat: 0, Lon: 0, Distance: 100, Stocked: 100, Delivered: 0},
		&db.Flyer{OwnerID: 1, Title: "title01", ValidPeriod: 3600, Lat: 0, Lon: 0, Distance: 100, Stocked: 100, Delivered: 0},
		&db.Flyer{OwnerID: 1, Title: "title01", ValidPeriod: 3600, Lat: 0, Lon: 0, Distance: 100, Stocked: 100, Delivered: 0},
	}
	locations := []*db.Location{
		&db.Location{ID: 0, Lat: 0, Lon: 0, Time: now},
		&db.Location{ID: 1, Lat: 0, Lon: 0, Time: now},
		&db.Location{ID: 2, Lat: 0, Lon: 0, Time: now},
		&db.Location{ID: 3, Lat: 2, Lon: 2, Time: now},
		&db.Location{ID: 4, Lat: 3, Lon: 3, Time: now},
	}

	// Start Server
	srv := newServer()
	srv.addr = "localhost:8081"
	srv.routes()
	go http.ListenAndServe(srv.addr, srv.router)
	// must wait here??

	// Post flyers
	for _, flyer := range flyers {
		flyerJson, err := json.Marshal(flyer)
		if err != nil {
			t.Fatalf("Flyer Marshal error: %v\n", err)
		}
		reqBody := flyerJson
		req, err := http.NewRequest("POST", "http://"+srv.addr+"/api/flyers", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Length", fmt.Sprintf("%v", len(reqBody)))
		if err != nil {
			t.Errorf("failed http.NewRequest %v", err)
		}
		client := new(http.Client)
		r, err := client.Do(req)
		if err != nil {
			t.Errorf("failed http.Client.Do %v", err)
		}
		defer r.Body.Close()

		if r.StatusCode != http.StatusOK {
			fmt.Printf("result:%#v\n", r)
			t.Errorf("get StatusCode:%v", r.StatusCode)
		}
	}

	// Post Locations
	for _, loc := range locations {
		time.Sleep(2 * time.Millisecond)
		err := loc.Set()
		if err != nil {
			t.Errorf("Set Location error: %v\n", err)
		}
	}

	time.Sleep(500 * time.Millisecond)
	//DELETE TEST DATA
	defer func() {
	}()
	db.DropFlyer()
	db.DropNotification()
	log.SetOutput(os.Stdout)
}

func TestWebhook(t *testing.T) {
	// No log
	//log.SetOutput(ioutil.Discard)

	now := time.Now().Unix()
	geometry := &db.Geometry{
		Type:        "Point",
		Coordinates: []byte(fmt.Sprintf("[%v,%v]", 1.2345, 100.2003)),
	}
	/*
		geometry_json, err := json.Marshal(geometry)
		if err != nil {
			t.Fatalf("JSON Marshal error: %v", err)
		}
	*/

	feature := &db.GeoJsonFeature{Type: "Feature", Geometry: geometry, Properties: map[string]interface{}{"id": float64(1), "time": float64(now)}}

	feature_json, err := json.Marshal(feature)
	if err != nil {
		t.Fatalf("JSON Marshal error: %v", err)
	}
	fmt.Printf("feature_json=%s\n", feature_json)

	requestData := []*WebhookRequest{
		// requestData[0] normal
		&WebhookRequest{Command: "set", Group: "", Detect: "enter", Hook: "flyerhook:1", Key: "location", Id: "1:1", Object: feature_json},
		// requestData[1] bad Hook
		&WebhookRequest{Command: "set", Group: "", Detect: "enter", Hook: "XXXXXXXXXXX", Key: "location", Id: "1:1", Object: feature_json},
		// requestData[2] flyerhook not exist
		&WebhookRequest{Command: "set", Group: "", Detect: "enter", Hook: "flyerhook:10000", Key: "location", Id: "1:1", Object: feature_json},
		// requestData[3] no flyer stock
		&WebhookRequest{Command: "set", Group: "", Detect: "enter", Hook: "flyerhook:2", Key: "location", Id: "1:1", Object: feature_json},
	}

	flyerData := []*db.Flyer{
		// flyerID[0] normal
		&db.Flyer{ID: 1, OwnerID: 1, Title: "title01", ValidPeriod: 3600, StartAt: now, EndAt: now + 3600, Lat: 0, Lon: 0, Distance: 100, Stocked: 100, Delivered: 0},
		// flyerID[0] no flyer stock
		&db.Flyer{ID: 2, OwnerID: 1, Title: "title02", ValidPeriod: 3600, StartAt: now, EndAt: now + 3600, Lat: 0, Lon: 0, Distance: 100, Stocked: 0, Delivered: 0},
	}

	tests := []struct {
		request *WebhookRequest
		method  string
		url     string
		flyer   *db.Flyer
		status  int
	}{
		// tests[0] normal
		{method: "POST", request: requestData[0], url: "/hook/notification", status: 200},
		// tests[1] wrong http method
		{method: "GET", request: requestData[0], url: "/hook/notification", status: 500},
		// tests[2] wrong hook ID
		{method: "POST", request: requestData[1], url: "/hook/notification", status: 500},
		// tests[3] flyer does not exist
		{method: "POST", request: requestData[2], url: "/hook/notification", status: 500},
		// tests[4] no flyer stock
		{method: "POST", request: requestData[3], url: "/hook/notification", status: 500},
	}

	// SET DATA
	for _, flyer := range flyerData {
		if err := flyer.Set(); err != nil {
			t.Fatalf("Set Flyer error: (%v) flyer:%v\n", err, flyer)
		}
		fmt.Printf("flyer.ID=%v\n", flyer.ID)
	}

	// Start Server
	srv := newServer()
	srv.routes()
	go http.ListenAndServe(srv.addr, srv.router)

	// TEST
	for test_number, test := range tests {
		if test.method == http.MethodGet {
			test.url = fmt.Sprintf("%v?", test.url)
		}
		reqBody, err := json.Marshal(test.request)
		if err != nil {
			t.Fatalf("request JSON Marshal error: %v", err)
		}
		req, err := http.NewRequest(test.method, "http://"+srv.addr+test.url, bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Length", fmt.Sprintf("%v", len(reqBody)))
		if err != nil {
			t.Errorf("tests[%v] failed http.NewRequest %v", test_number, err)
		}
		client := new(http.Client)
		r, err := client.Do(req)
		if err != nil {
			t.Errorf("failed http.Client.Do %v", err)
		}
		defer r.Body.Close()

		if test.status == 0 && r.StatusCode != http.StatusOK ||
			test.status != 0 && r.StatusCode != test.status {
			fmt.Printf("result:%#v\n", r)
			t.Errorf("tests[%v] method:%v url:%v StatusCode:%v", test_number, test.method, test.url, r.StatusCode)
		}
		if test.status != 0 {
			continue
		}
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("result:%#v\n", r)
			t.Errorf("tests[%v] method:%v url:%v Error by ioutil.ReadAll(). %v", test_number, test.method, test.url, err)
		}
		log.Printf("Body:%v\n", string(data))
		db.DropNotification()
	}

	//DELETE TEST DATA
	db.DropFlyer()
	log.SetOutput(os.Stdout)
}
