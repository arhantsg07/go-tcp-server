// This file contains the server definition struct and all client-management operations.

package server

import (
	"fmt"
	"log"
	"sync"

	"github.com/arhantsg07/go-tcp-server/internal/protocol"
)

type Server struct {
	ip      string
	port    string
	clients map[string]*Connection
	mu      sync.RWMutex // RWMutex: many concurrent reads (broadcast), rare writes (join/leave)
	metrics *Metrics
}

func New(ip, port string) *Server {
	return &Server{
		ip:      ip,
		port:    port,
		clients: make(map[string]*Connection),
		metrics: newMetrics(),
	}
}

// Register adds a client to the registry under the given username.
// Returns an error if the username is already taken.
func (s *Server) Register(username string, conn *Connection) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.clients[username]; exists {
		return fmt.Errorf("username %q is already taken", username)
	}
	conn.username = username
	s.clients[username] = conn
	s.metrics.TotalConnections.Add(1)
	s.metrics.ActiveConnections.Add(1)
	return nil
}

// Unregister removes a client from the registry. Safe to call if username is not registered.
func (s *Server) Unregister(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.clients[username]; exists {
		delete(s.clients, username)
		s.metrics.ActiveConnections.Add(-1)
	}
}

// Broadcast sends a message to every connected client except the sender.
// It collects targets under a read lock, then releases before sending to avoid
// holding the lock during potentially slow network writes.
func (s *Server) Broadcast(msg protocol.Message) {
	s.mu.RLock()
	targets := make([]*Connection, 0, len(s.clients))
	for username, conn := range s.clients {
		if username != msg.From {
			targets = append(targets, conn)
		}
	}
	s.mu.RUnlock()

	for _, conn := range targets {
		if err := conn.SendMessage(msg); err != nil {
			log.Printf("broadcast error to %s: %v", conn.username, err)
		} else {
			s.metrics.MessagesSent.Add(1)
			s.metrics.BroadcastsSent.Add(1)
		}
	}
}

// SendPrivate delivers a message to a single named client.
func (s *Server) SendPrivate(to string, msg protocol.Message) error {
	s.mu.RLock()
	conn, exists := s.clients[to]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("user %q not found", to)
	}
	if err := conn.SendMessage(msg); err != nil {
		return err
	}
	s.metrics.MessagesSent.Add(1)
	s.metrics.PrivatesSent.Add(1)
	return nil
}

// OnlineUsers returns a snapshot of currently connected usernames.
func (s *Server) OnlineUsers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	users := make([]string, 0, len(s.clients))
	for u := range s.clients {
		users = append(users, u)
	}
	return users
}

// Metrics returns the server's live metrics collector.
func (s *Server) Metrics() *Metrics {
	return s.metrics
}
