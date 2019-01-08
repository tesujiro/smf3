package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
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
	s := newServer()
	s.routes()
	http.ListenAndServe("localhost:8000", s.router)
}

func (s *server) routes() {
	s.router.HandleFunc("/", s.handleDefault())
	s.router.HandleFunc("/greet", s.handleHello())
	s.router.HandleFunc("/manual", s.manual())
	s.router.HandleFunc("/location", s.handleLocation())
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
		http.Redirect(w, r, "/manual", 301)
	}
}

func (s *server) manual() http.HandlerFunc {
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

func pseudoClient() string {
	dogenzaka := []jsonmap{
		{"lat": 35.6591536, "lon": 139.69818840000002},
		{"lat": 35.6591365, "lon": 139.69812520000002},
		{"lat": 35.6591022, "lon": 139.697998},
		{"lat": 35.6590627, "lon": 139.6978704},
		{"lat": 35.6590219, "lon": 139.6977639},
		{"lat": 35.659004200000005, "lon": 139.6977287},
		{"lat": 35.6589497, "lon": 139.6976223},
		{"lat": 35.6588903, "lon": 139.6975169},
		{"lat": 35.658733000000005, "lon": 139.6973049},
		{"lat": 35.6586986, "lon": 139.69726070000002},
		{"lat": 35.6585612, "lon": 139.6970843},
		{"lat": 35.6584173, "lon": 139.696921},
		{"lat": 35.658166800000004, "lon": 139.696693},
		{"lat": 35.6578986, "lon": 139.69648600000002},
		{"lat": 35.6578357, "lon": 139.69643480000002},
		{"lat": 35.6576721, "lon": 139.69630170000002},
		{"lat": 35.6575015, "lon": 139.69616290000002},
		{"lat": 35.6573663, "lon": 139.69605380000002},
		{"lat": 35.657306600000005, "lon": 139.69601360000001},
	}
	list := []jsonmap{}
	for i := 0; i < 3; i++ {
		list = append(list, dogenzaka[rand.Intn(len(dogenzaka))])
	}
	m := jsonmap{"list": list}
	j, err := json.Marshal(m)
	if err != nil {
		fmt.Printf("json.Marshall error:%v\n", err)
	}
	return fmt.Sprintf(string(j))
}

func (s *server) handleLocation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Location Received:")

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
			log.Printf("read failed!!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("Content-Length:%v", length)
		log.Printf("Content-Body:%s", body)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, pseudoClient())
	}
}
