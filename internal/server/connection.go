// The connection hnadler is moved here from tcp.go

package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

type Connection struct {
	conn   net.Conn
	reader *bufio.Reader
}

func NewConn(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
		reader: bufio.NewReader(conn),
	}
}

func HandleConnection(conn *Connection) {

	defer conn.conn.Close()

	// buffer := make([]byte, 1024)

	for {

		// READ

		rcv_msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("The following error was encountered: ", err)
		} else {
			fmt.Println("received: ", rcv_msg)
		}

		// WRITE
		msg := Message{
			time.Now(),
			"Hello, how are you ?",
		}

		err = conn.SendMessage(msg)
		if err != nil {
			log.Println("Encountered write error: ", err)
		}

	}
}
