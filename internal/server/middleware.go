// middleware.go - decorator-style middleware for the MessageHandler interface.

package server

import (
	"log"
	"time"

	"github.com/arhantsg07/go-tcp-server/internal/protocol"
)

// MiddlewareFunc is a function that wraps a MessageHandler with additional behaviour.
type MiddlewareFunc func(MessageHandler) MessageHandler

// LoggingMiddleware wraps a MessageHandler and logs every read and write
// along with the operation latency.
func LoggingMiddleware(next MessageHandler) MessageHandler {
	return &loggingHandler{next: next}
}

type loggingHandler struct {
	next MessageHandler
}

func (l *loggingHandler) ReadMessage() (protocol.Message, error) {
	start := time.Now()
	msg, err := l.next.ReadMessage()
	if err == nil {
		log.Printf("[READ]  type=%-12s from=%-15s latency=%s",
			msg.Type, msg.From, time.Since(start))
	}
	return msg, err
}

func (l *loggingHandler) SendMessage(msg protocol.Message) error {
	start := time.Now()
	err := l.next.SendMessage(msg)
	if err == nil {
		log.Printf("[SEND]  type=%-12s to=%-15s  latency=%s",
			msg.Type, msg.To, time.Since(start))
	}
	return err
}
