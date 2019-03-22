package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tesujiro/smf3/data/db"
	"github.com/tesujiro/smf3/debug"
)

func TestAPINotification(t *testing.T) {
	db.DropNotification()

	now := time.Now().Unix()
	notifs := []*db.Notification{
		&db.Notification{ID: "0:0", Lat: 0, Lon: 0, DeliveryTime: now},
		&db.Notification{ID: "1:1", Lat: 1, Lon: 1, DeliveryTime: now},
		&db.Notification{ID: "2:2", Lat: 2, Lon: 2, DeliveryTime: now},
		&db.Notification{ID: "3:3", Lat: 3, Lon: 3, DeliveryTime: now},
	}
	data := []*db.Notification{notifs[0], notifs[1], notifs[2]}
	tests := []struct {
		method         string
		url            string
		bounds         map[string]string
		contentType    string
		notification   *db.Notification
		body           string
		status         int
		header         map[string]string
		expectedLength int
	}{
		{method: "GET", url: "/api/notifications", bounds: map[string]string{"south": "0", "west": "0", "north": "1", "EAST": "1"}, expectedLength: 2},
		{method: "GET", url: "/api/notifications", contentType: "text/plain", bounds: map[string]string{"south": "0", "west": "0", "north": "1", "east": "1"}, status: 500},
		{method: "GET", url: "/api/notifications", bounds: map[string]string{}, status: 500},
		{method: "GET", url: "/api/notifications", bounds: map[string]string{"xxx": "0"}, status: 500},
		{method: "GET", url: "/api/notifications", bounds: map[string]string{"south": "0", "west": "0", "north": "0", "east": "xxx"}, status: 500},
		{method: "GET", url: "/api/notifications", bounds: map[string]string{"south": "0", "west": "0", "north": "0"}, status: 500},
		{method: "GET", url: "/api/notifications", bounds: map[string]string{"south": "1", "SOUTH": "1", "west": "0", "north": "1", "EAST": "1"}, status: 500},
		//{method: "POST", url: "/api/notifications", notification: notifs[2], expectedLength: 3},
	}

	// SET DATA
	for _, notification := range data {
		if err := notification.Set(); err != nil {
			t.Fatalf("Set Notification error: (%v) notification:%v\n", err, notification)
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
			if reqBody, err = json.Marshal(test.notification); err != nil {
				t.Errorf("Test[%v] failed to marshal request body: %v", test_number, err)
			}
		}
		req, err := http.NewRequest(test.method, test.url, bytes.NewReader(reqBody))
		if test.contentType != "" {
			req.Header.Set("Content-Type", test.contentType)
		} else {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Content-Length", fmt.Sprintf("%v", len(reqBody)))
		if err != nil {
			t.Errorf("Test[%v] failed http.NewRequest %v", test_number, err)
		}
		w := httptest.NewRecorder()
		srv.router.ServeHTTP(w, req)
		r := w.Result()
		debug.Printf("Result:%#v\n", r)
		if test.status == 0 && r.StatusCode != http.StatusOK ||
			test.status != 0 && r.StatusCode != test.status {
			fmt.Printf("result:%#v\n", r)
			t.Errorf("Test[%v] method:%v url:%v StatusCode:%v", test_number, test.method, test.url, r.StatusCode)
		}
		if test.status != 0 {
			continue
		}
		//fmt.Printf("header.Notification:%#v\n", r.Header["Notification"])
		switch test.method {
		case http.MethodPost:

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
			debug.Printf("Body:%v\n", string(data))
			var actualNotifications []*db.Notification
			if err := json.Unmarshal(data, &actualNotifications); err != nil {
				t.Errorf("Test[%v] Response body json.Unmarshal error: %v", test_number, err)
			}
			if len(actualNotifications) != test.expectedLength {
				t.Errorf("Test[%v] Response length error. expected: %v actual: %v", test_number, test.expectedLength, len(actualNotifications))
			}
		}
	}

	//DELETE TEST DATA
	db.DropNotification()
}
