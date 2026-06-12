// stats_server.go - lightweight HTTP server exposing live server metrics.

package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// StartStatsServer starts an HTTP server on the given port.
// Three endpoints are available:
//
//	GET /stats   — full JSON metrics snapshot
//	GET /clients — list of connected usernames
//	GET /health  — plain-text liveness probe
func (s *Server) StartStatsServer(port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(s.metrics.Snapshot()); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/clients", func(w http.ResponseWriter, r *http.Request) {
		users := s.OnlineUsers()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"count": len(users),
			"users": users,
		})
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})

	addr := ":" + port
	log.Printf("stats server listening on http://localhost%s  (endpoints: /stats /clients /health)", addr)
	if err := http.ListenAndServe(addr, mux); err != nil && err != http.ErrServerClosed {
		log.Printf("stats server error: %v", err)
	}
}
