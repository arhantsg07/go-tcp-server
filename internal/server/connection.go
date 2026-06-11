// connection.go - Connection type, per-connection lifecycle, and message routing.

package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/arhantsg07/go-tcp-server/internal/protocol"
)

// Connection wraps a raw net.Conn with buffered reading, a per-connection
// write mutex, a rate limiter, and a reference to server-wide metrics.
type Connection struct {
	conn     net.Conn
	reader   *bufio.Reader
	username string
	writeMu  sync.Mutex // serialises concurrent writes (heartbeat goroutine + broadcast)
	limiter  *RateLimiter
	metrics  *Metrics
}

// NewConn creates a Connection ready for use.
func NewConn(conn net.Conn, metrics *Metrics) *Connection {
	return &Connection{
		conn:    conn,
		reader:  bufio.NewReader(conn),
		limiter: NewRateLimiter(10, 5), // burst=10, sustained=5 msg/s
		metrics: metrics,
	}
}

// HandleConnection owns the full lifecycle of a single client connection:
// welcome → JOIN handshake → message loop → cleanup.
func (s *Server) HandleConnection(conn *Connection) {
	// Cleanup runs on every exit path: close socket, remove from registry,
	// broadcast departure to remaining clients.
	defer func() {
		conn.conn.Close()
		if conn.username != "" {
			s.Unregister(conn.username)
			s.Broadcast(protocol.Message{
				Type:      protocol.MessageTypeChat,
				From:      "Server",
				Timestamp: time.Now(),
				Text:      fmt.Sprintf("%s has left the chat.", conn.username),
			})
			log.Printf("disconnected: %s (%s)", conn.username, conn.conn.RemoteAddr())
		}
	}()

	// ── Step 1: prompt the client to identify themselves ───────────────────
	if err := conn.SendMessage(protocol.Message{
		Type:      protocol.MessageTypeJoin,
		From:      "Server",
		Timestamp: time.Now(),
		Text:      "Connected. Send a JOIN message with your username to continue.",
	}); err != nil {
		log.Println("welcome send failed:", err)
		return
	}

	// ── Step 2: wait for JOIN with a username (15-second deadline) ─────────
	conn.conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	joinMsg, err := conn.ReadMessage()
	if err != nil || joinMsg.Type != protocol.MessageTypeJoin || strings.TrimSpace(joinMsg.From) == "" {
		conn.SendMessage(protocol.Message{
			Type: protocol.MessageTypeError, From: "Server",
			Text: "Expected a JOIN message with a non-empty 'from' field. Closing.",
		})
		return
	}

	if err := s.Register(joinMsg.From, conn); err != nil {
		conn.SendMessage(protocol.Message{
			Type: protocol.MessageTypeError, From: "Server",
			Text: err.Error(),
		})
		return
	}

	// Confirm join to the new client, then announce their arrival to everyone else.
	conn.SendMessage(protocol.Message{
		Type:      protocol.MessageTypeJoin,
		From:      "Server",
		To:        conn.username,
		Timestamp: time.Now(),
		Text:      fmt.Sprintf("Welcome, %s! %d user(s) online.", conn.username, len(s.OnlineUsers())),
	})
	s.Broadcast(protocol.Message{
		Type:      protocol.MessageTypeChat,
		From:      "Server",
		Timestamp: time.Now(),
		Text:      fmt.Sprintf("%s has joined the chat.", conn.username),
	})
	log.Printf("registered: %s (%s)", conn.username, conn.conn.RemoteAddr())

	// ── Step 3: start heartbeat keep-alive ─────────────────────────────────
	stopHeartbeat := s.startHeartbeat(conn)
	defer close(stopHeartbeat)

	// ── Step 4: main message loop ───────────────────────────────────────────
	// Wrap with LoggingMiddleware so every read and write is timestamped.
	var handler MessageHandler = conn
	handler = LoggingMiddleware(handler)

	for {
		msg, err := handler.ReadMessage()
		if err != nil {
			return
		}

		s.metrics.MessagesReceived.Add(1)

		if !conn.limiter.Allow() {
			handler.SendMessage(protocol.Message{
				Type:      protocol.MessageTypeError,
				From:      "Server",
				Timestamp: time.Now(),
				Text:      "Rate limit exceeded — slow down.",
			})
			continue
		}

		if !s.route(handler, conn, msg) {
			return
		}
	}
}

// route dispatches msg to the correct handler based on its type.
// Returns false when the connection should be terminated.
func (s *Server) route(handler MessageHandler, conn *Connection, msg protocol.Message) bool {
	// Enforce server-side identity — clients cannot forge the From field.
	msg.From = conn.username
	msg.Timestamp = time.Now()

	switch msg.Type {
	case protocol.MessageTypeChat:
		s.Broadcast(msg)

	case protocol.MessageTypePrivate:
		if msg.To == "" {
			handler.SendMessage(protocol.Message{
				Type: protocol.MessageTypeError, From: "Server",
				Timestamp: time.Now(), Text: "PRIVATE requires a non-empty 'to' field.",
			})
			return true
		}
		if err := s.SendPrivate(msg.To, msg); err != nil {
			handler.SendMessage(protocol.Message{
				Type: protocol.MessageTypeError, From: "Server",
				Timestamp: time.Now(), Text: err.Error(),
			})
		}

	case protocol.MessageTypeList:
		users := s.OnlineUsers()
		handler.SendMessage(protocol.Message{
			Type:      protocol.MessageTypeList,
			From:      "Server",
			To:        conn.username,
			Timestamp: time.Now(),
			Text:      strings.Join(users, ", "),
		})

	case protocol.MessageTypeHeartbeat:
		// Client echoing our heartbeat back — no-op. The read itself proves liveness.

	case protocol.MessageTypeDisconnect:
		handler.SendMessage(protocol.Message{
			Type:      protocol.MessageTypeDisconnect,
			From:      "Server",
			Timestamp: time.Now(),
			Text:      fmt.Sprintf("Goodbye, %s!", conn.username),
		})
		return false

	default:
		log.Printf("unknown message type %q from %s", msg.Type, conn.username)
	}
	return true
}
