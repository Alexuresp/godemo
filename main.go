package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

//go:embed templates/*
var templateFS embed.FS

type entry struct {
	Name      string
	Message   string
	Timestamp time.Time
}

type server struct {
	tmpl    *template.Template
	mu      sync.Mutex
	entries []entry
}

func main() {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		log.Fatalf("parse templates: %v", err)
	}

	s := &server{tmpl: tmpl}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /", s.index)
	mux.HandleFunc("POST /", s.submit)

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func (s *server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	entries := append([]entry(nil), s.entries...)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "index.html", entries); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *server) submit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	message := r.FormValue("message")
	if name == "" || message == "" {
		http.Error(w, "name and message are required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.entries = append([]entry{{
		Name:      name,
		Message:   message,
		Timestamp: time.Now().UTC(),
	}}, s.entries...)
	if len(s.entries) > 50 {
		s.entries = s.entries[:50]
	}
	s.mu.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
