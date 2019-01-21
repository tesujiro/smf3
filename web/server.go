package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/tesujiro/smf3/data/db"
	"github.com/tesujiro/smf3/match"
)

type server struct {
	router *http.ServeMux
}

func newServer() *server {
	return &server{
		router: http.NewServeMux(),
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// DELETE DATA
	db.DropFlyer()
	db.DropNotification()

	// START MATCHING ENGINE
	matcher := match.NewMatcher(ctx)
	if err := matcher.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return
	}

	// START WEB SERVER
	s := newServer()
	s.routes()
	http.ListenAndServe("localhost:8000", s.router)

	<-ctx.Done()
}

func (s *server) routes() {
	s.router.HandleFunc("/", s.handleDefault())
	s.router.HandleFunc("/greet", s.handleHello())
	s.router.HandleFunc("/portal", s.portal())
	s.router.HandleFunc("/location", s.handleLocation())
	s.router.HandleFunc("/footway", s.handleFootway()) // TODO: /api/footways
	s.router.HandleFunc("/api/locations", s.handleLocations())
	//s.router.HandleFunc("/api/locations/", s.handleSingleLocation())
	s.router.HandleFunc("/api/flyers", s.handleFlyers())
	//s.router.HandleFunc("/api/flyers/", s.handleSingleFlyer())
	//s.router.HandleFunc("/api/notifications", s.handleNotifs())
	//s.router.HandleFunc("/api/notifications/", s.handleSingleNotifs())
	s.router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./public"))))
}

func (s *server) handleHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World")
	}
}

func (s *server) handleDefault() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Default Handler!!")
		log.Printf("URL=%v\n", r.URL)
		http.Redirect(w, r, "/portal", 301)
	}
}

func (s *server) portal() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Consumer Manual Tester Page!!")
		tpl := template.Must(template.ParseFiles("template/ManualTester.html"))
		w.Header().Set("Content-Type", "text/html")

		err := tpl.Execute(w, map[string]string{"APIKEY": os.Getenv("APIKEY")})
		if err != nil {
			panic(err)
		}
	}
}

type jsonmap map[string]interface{}

func getFootway() ([]byte, error) {
	path := "../data/osm/ways_on_browser.json"
	return ioutil.ReadFile(path)
}

func (s *server) handleLocation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//log.Printf("Location Received:")

		if r.Header.Get("Content-Type") != "application/json" {
			log.Printf("bad Content-Type!!")
			log.Printf(r.Header.Get("Content-Type"))
		}

		//To allocate slice for request body
		length, err := strconv.Atoi(r.Header.Get("Content-Length"))
		if err != nil {
			log.Printf("Content-Length failed!!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		//Read body data to parse json
		body := make([]byte, length)
		length, err = r.Body.Read(body)
		if err != nil && err != io.EOF {
			log.Printf("read failed!!\n")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		//log.Printf("Content-Length:%v", length)
		//log.Printf("Content-Body:%s", body)

		type Req struct {
			Bounds map[string]float64 `json:"bounds"`
			Flyers []db.Flyer         `json:"flyers"`
			//Flyers []map[string]interface{} `json:"flyers"`
		}
		type Flyer struct {
		}
		var reqInfo Req
		if err := json.Unmarshal(body, &reqInfo); err != nil {
			log.Printf("Request body marshaling  error: %v\n", err)
			return
		}

		bounds := reqInfo.Bounds
		//fmt.Printf("request.bounds=%v\n", bounds)

		for _, f := range reqInfo.Flyers {
			now := time.Now().Unix()
			f.ID = now //TODO: temporary
			f.StartAt = now
			f.EndAt = now + f.ValidPeriod
			if err := f.Set(); err != nil {
				log.Printf("Set Flyer error: (%v) flyer:%v\n", err, f)
				return
			}
		}

		// ----------------------------------------------------------------------------------------
		// MAKE REAPONSE DATA

		// 3. notifications
		var notifications []db.GeoJsonFeature
		var notificationJson []byte
		notifications, err = db.NotificationWithinBounds(bounds["south"], bounds["west"], bounds["north"], bounds["east"])
		if err != nil {
			log.Printf("WithiLocation error: %v\n", err)
			return
		}
		notificationJson, err = json.Marshal(notifications)
		if err != nil {
			log.Printf("Notification Marshal error: %v\n", err)
			return
		}
		//fmt.Fprintf(w, "%s", notificationJson)

		// write json data
		fmt.Fprintf(w, `{"notifications": %s}`, notificationJson)
		//fmt.Fprintf(w, `{"locations": %s,"flyers": %s, "notifications": %s}`, locationJson, flyerJson, notificationJson)

		fmt.Printf("request: {bounds: %v ,flyers: %v }\t", len(reqInfo.Bounds), len(reqInfo.Flyers))
		//fmt.Printf("response: {locations: %v ,flyers: %v , notifications: %v }\n", len(locations), len(flyers), len(notifications))
		fmt.Printf("response: {notifications: %v }\n", len(notifications))
	}
}
