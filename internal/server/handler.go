// moved the message handling (read / write) from connection.go to handler.go

package server

import (
	"fmt"
	"io"
	"net"
	"time"
)

// declared the two functions as the part of ths msg handling interface
// these methods are implemented by the net.Conn type

type MessageHandler interface {
	ReadMessage(buffer []byte) (int, error)
	SendMessage(msg Message) error
}

type Message struct {
	Timestamp time.Time
	Text      string
}

func (c *Connection) ReadMessage(buffer []byte) (int, error) {
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	n, err := c.conn.Read(buffer)

	if err != nil {
		if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
			// fmt.Println("client timed out: ", nErr)
			return n, nErr
		}

		if err == io.EOF {
			fmt.Println("client disconnected")
			return n, err
		}

		return n, err
	}
	// successfully read/ complete, no error
	return n, nil
}

func (c *Connection) SendMessage(msg Message) error {
	c.conn.SetWriteDeadline(time.Now().Add(30 * time.Second))


	formatted := fmt.Sprintf("From Server (%s) : %s\n", msg.Timestamp.Format(time.RFC3339), msg.Text)
	_, err := c.conn.Write([]byte(formatted))
	if err != nil {
		fmt.Println("Write error: ", err)
		return err
	}
	return nil
}
