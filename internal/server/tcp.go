package server

import (
	"context"
	"fmt"
	"log"
	"net"
)

// Start begins accepting connections on s.ip:s.port.
// It blocks until ctx is cancelled, at which point it closes the listener and
// returns nil — allowing the caller to treat a clean shutdown as a non-error.
func (s *Server) Start(ctx context.Context) error {
	addr := s.ip + ":" + s.port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	log.Printf("TCP server listening on %s", addr)

	// Close the listener when the context is cancelled. This unblocks Accept
	// below and lets the loop detect the shutdown signal.
	go func() {
		<-ctx.Done()
		log.Println("shutdown signal received — closing listener")
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				log.Println("server shut down gracefully")
				return nil
			default:
				log.Println("accept error:", err)
				continue
			}
		}
		log.Printf("new connection from %s", conn.RemoteAddr())
		newConn := NewConn(conn, s.metrics)
		go s.HandleConnection(newConn)
	}
}
