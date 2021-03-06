package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/tesujiro/smf3/data/db"
)

type server struct {
	router *http.ServeMux
	addr   string
}

func newServer() *server {
	addr := "localhost:8000"
	return &server{
		router: http.NewServeMux(),
		addr:   addr,
	}
}

func (s *server) notificationEndpoint() string {
	return "http://" + s.addr + "/hook/notification"
}

func main() {
	log.SetFlags(log.Lmicroseconds)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// DELETE DATA
	db.DropFlyer()
	db.DropNotification()

	// START WEB SERVER
	s := newServer()
	s.routes()
	http.ListenAndServe(s.addr, s.router)

	<-ctx.Done()
}

func (s *server) routes() {
	s.router.HandleFunc("/", s.handleDefault())
	s.router.HandleFunc("/greet", s.handleHello())
	s.router.HandleFunc("/portal", s.portal())
	s.router.HandleFunc("/footway", s.handleFootway()) // TODO: /api/footways
	s.router.HandleFunc("/api/locations", s.handleLocations())
	//s.router.HandleFunc("/api/locations/", s.handleSingleLocation())
	s.router.HandleFunc("/api/flyers", s.handleFlyers())
	//s.router.HandleFunc("/api/flyers/", s.handleSingleFlyer())
	s.router.HandleFunc("/api/notifications", s.handleNotifications())
	// Webhook
	s.router.HandleFunc("/hook/notification", s.hookNotifications())
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
		tpl := template.Must(template.ParseFiles("./template/ManualTester.html"))
		w.Header().Set("Content-Type", "text/html")

		err := tpl.Execute(w, map[string]string{"APIKEY": os.Getenv("APIKEY")})
		if err != nil {
			panic(err)
		}
	}
}

type jsonmap map[string]interface{}
