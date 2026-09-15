package main

import (
	"backend/api"
	"log"
	"net/http"
	"os"
	"time"
)

func newHTTPServer(address string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("LingoLift API listening on http://localhost:%s", port)
	log.Fatal(newHTTPServer(":"+port, api.NewHandler()).ListenAndServe())
}
