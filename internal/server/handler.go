// moved the message handling (read / write) from connection.go to handler.go

package server

import (
	"fmt"
	"io"
	"net"
	"time"
	"github.com/arhantsg07/go-tcp-server/internal/protocol"
)

// adding message type to understand the type of communication

// declared the two functions as the part of ths msg handling interface
// these methods are implemented by the net.Conn type

type MessageHandler interface {
	ReadMessage() (protocol.Message, error)
	SendMessage(msg protocol.Message) error
}

func (c *Connection) ReadMessage() (protocol.Message, error) {
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	raw, err := c.reader.ReadBytes('\n')
	if err != nil {
		if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
			return protocol.Message{}, nErr
		}
		if err == io.EOF {
			fmt.Println("client disconnected")
			return protocol.Message{}, err
		}
		return protocol.Message{}, err
	}

	data, err := protocol.Decode(raw)
	if err != nil {
		return protocol.Message{}, fmt.Errorf("decode error: %w", err)
	}

	// successfully read and decoded
	return data, nil
}

func (c *Connection) SendMessage(msg protocol.Message) error {
	c.conn.SetWriteDeadline(time.Now().Add(30 * time.Second))

	// formatted := fmt.Sprintf("From %s to %s (%s) : %s\n", msg.From, msg.To, msg.Timestamp.Format(time.RFC3339), msg.Text)
	data, err := protocol.Encode(msg)
	if err != nil {
		return err
	}

	_, err = c.conn.Write(data)
	if err != nil {
		fmt.Println("Write error: ", err)
		return err
	}
	return nil
}
