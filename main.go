package main

import (
	"context"
	"database/sql"
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

//go:embed templates/*
var templateFS embed.FS

type entry struct {
	Name      string
	Message   string
	Timestamp time.Time
}

type pageData struct {
	Entries []entry
	Storage string
}

type server struct {
	tmpl *template.Template
	db   *sql.DB
}

func main() {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		log.Fatalf("parse templates: %v", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := waitForDB(ctx, db); err != nil {
		log.Fatalf("db not ready: %v", err)
	}
	if err := migrate(ctx, db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	s := &server{tmpl: tmpl, db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /", s.index)
	mux.HandleFunc("POST /", s.submit)

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	log.Printf("listening on %s (postgres)", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func waitForDB(ctx context.Context, db *sql.DB) error {
	var last error
	for {
		last = db.PingContext(ctx)
		if last == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return last
		case <-time.After(2 * time.Second):
		}
	}
}

func migrate(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS messages (
			id         BIGSERIAL PRIMARY KEY,
			name       TEXT NOT NULL,
			message    TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func (s *server) healthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.db.PingContext(ctx); err != nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `
		SELECT name, message, created_at
		FROM messages
		ORDER BY created_at DESC
		LIMIT 50
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var entries []entry
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.Name, &e.Message, &e.Timestamp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "index.html", pageData{
		Entries: entries,
		Storage: "PostgreSQL",
	}); err != nil {
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

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO messages (name, message) VALUES ($1, $2)`,
		name, message,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
