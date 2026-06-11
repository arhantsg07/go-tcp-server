// handler.go - low-level read/write methods on Connection, and the MessageHandler interface.

package server

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/arhantsg07/go-tcp-server/internal/protocol"
)

// MessageHandler is the interface for reading and writing protocol messages.
// Implementing it as an interface enables the middleware decorator pattern.
type MessageHandler interface {
	ReadMessage() (protocol.Message, error)
	SendMessage(msg protocol.Message) error
}

// ReadMessage reads the next newline-delimited JSON frame from the connection.
// It enforces a 30-second read deadline to detect stale/dead clients.
func (c *Connection) ReadMessage() (protocol.Message, error) {
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	raw, err := c.reader.ReadBytes('\n')
	if err != nil {
		if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
			return protocol.Message{}, nErr
		}
		if err == io.EOF {
			return protocol.Message{}, err
		}
		return protocol.Message{}, err
	}

	// Track raw bytes received for throughput metrics.
	if c.metrics != nil {
		c.metrics.BytesReceived.Add(int64(len(raw)))
	}

	msg, err := protocol.Decode(raw)
	if err != nil {
		return protocol.Message{}, fmt.Errorf("decode error: %w", err)
	}
	return msg, nil
}

// SendMessage encodes a Message as JSON and writes it to the connection.
// writeMu ensures only one goroutine writes at a time (heartbeat + broadcast can race).
func (c *Connection) SendMessage(msg protocol.Message) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	data, err := protocol.Encode(msg)
	if err != nil {
		return err
	}

	c.conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
	n, err := c.conn.Write(data)
	if err != nil {
		return err
	}

	if c.metrics != nil {
		c.metrics.BytesSent.Add(int64(n))
	}
	return nil
}
