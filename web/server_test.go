package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer(t *testing.T) {
	cases := []struct {
		method string
		url    string
		status int
		body   string
		header map[string]string
	}{
		{method: "GET", url: "/greet", body: "Hello, World"},
		{method: "GET", url: "/static", status: http.StatusMovedPermanently},
		{method: "GET", url: "/static/index.html", status: http.StatusMovedPermanently},
		{method: "GET", url: "/no_page", status: http.StatusMovedPermanently},
	}
	srv := newServer()
	srv.routes()
	for _, c := range cases {
		req, err := http.NewRequest(c.method, c.url, nil)
		if err != nil {
			t.Errorf("failed http.NewRequest %v", err)
		}
		w := httptest.NewRecorder()
		srv.router.ServeHTTP(w, req)
		r := w.Result()
		//fmt.Printf("Result:%#v\n", r)
		if c.status == 0 && r.StatusCode != http.StatusOK ||
			c.status != 0 && r.StatusCode != c.status {
			fmt.Printf("result:%#v\n", r)
			t.Errorf("method:%v url:%v StatusCode:%v", c.method, c.url, r.StatusCode)
		}
		//fmt.Printf("header.Location:%#v\n", r.Header["Location"])
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("result:%#v\n", r)
			t.Errorf("method:%v url:%v Error by ioutil.ReadAll(). %v", c.method, c.url, err)
		}
		if c.body != "" && string(data) != c.body {
			fmt.Printf("result:%#v\n", r)
			t.Errorf("method:%v url:%v Data Error. [%v]", c.method, c.url, string(data))
		}
	}
}
