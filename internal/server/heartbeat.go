// heartbeat.go - per-connection keep-alive goroutine.

package server

import (
	"log"
	"time"

	"github.com/arhantsg07/go-tcp-server/internal/protocol"
)

const (
	heartbeatInterval = 30 * time.Second
)

// startHeartbeat launches a background goroutine that sends a HEARTBEAT message
// to the client every heartbeatInterval. If the send fails (dead connection),
// the connection is closed so the main read loop unblocks and cleans up.
// The caller must close the returned channel to stop the goroutine.
func (s *Server) startHeartbeat(conn *Connection) chan struct{} {
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				err := conn.SendMessage(protocol.Message{
					Type:      protocol.MessageTypeHeartbeat,
					From:      "Server",
					Timestamp: time.Now(),
				})
				if err != nil {
					log.Printf("heartbeat failed for %s — closing connection", conn.username)
					conn.conn.Close() // unblocks ReadMessage in the main loop
					return
				}
			}
		}
	}()
	return stop
}
