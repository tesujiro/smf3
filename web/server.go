package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
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
	s.router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./public"))))
}

func (s *server) handleHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World")
	}
}

func (s *server) handleDefault() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
