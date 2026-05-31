// The connection hnadler is moved here from tcp.go

package server

import (
	"fmt"
	"log"
	"net"
	"time"
)


type Connection struct {
	conn net.Conn
}

func NewConn (conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
}


func HandleConnection(conn *Connection) {

	defer conn.conn.Close()

	buffer := make([]byte, 1024)

	for {
		
		// READ
		
		n, err := conn.ReadMessage(buffer)
		if err != nil {
			log.Println("The following error was encountered: ", err)
		} else {
			fmt.Println("received: ", string(buffer[:n]))
		}
		

		// WRITE
		msg := Message {
			time.Now(),
			"Hello, how are you ?",
		}

		err = conn.SendMessage(msg)
		if err != nil {
			log.Println("Encountered write error: ", err)
		}
		
	}
}
