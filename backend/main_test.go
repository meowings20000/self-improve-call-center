package main

import (
	"net/http"
	"testing"
	"time"
)

func TestHTTPServerHasDefensiveTimeouts(t *testing.T) {
	server := newHTTPServer(":8080", http.NewServeMux())

	want := map[string]struct {
		got  time.Duration
		want time.Duration
	}{
		"ReadHeaderTimeout": {server.ReadHeaderTimeout, 5 * time.Second},
		"ReadTimeout":       {server.ReadTimeout, 10 * time.Second},
		"WriteTimeout":      {server.WriteTimeout, 15 * time.Second},
		"IdleTimeout":       {server.IdleTimeout, 60 * time.Second},
	}
	for name, timeout := range want {
		if timeout.got != timeout.want {
			t.Errorf("%s = %s, want %s", name, timeout.got, timeout.want)
		}
	}
}
