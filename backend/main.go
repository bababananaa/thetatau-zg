package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"
)

// server holds the dependencies shared by all handlers.
// It is empty for now; the database handle will live here.
type server struct{}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s := &server{}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /api/people", notImplemented)
	mux.HandleFunc("POST /api/login", notImplemented)
	mux.HandleFunc("POST /api/logout", notImplemented)
	mux.HandleFunc("GET /api/me", notImplemented)
	mux.HandleFunc("GET /api/directory", notImplemented)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func notImplemented(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
