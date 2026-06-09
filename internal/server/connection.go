// The connection hnadler is moved here from tcp.go

package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
	"cmp"
	"github.com/arhantsg07/go-tcp-server/internal/protocol"
)

type Connection struct {
	conn   net.Conn
	reader *bufio.Reader
}

func NewConn(conn net.Conn) *Connection {
	return &Connection{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
}

func HandleConnection(conn *Connection) {

	defer conn.conn.Close()

	// Send welcome message once when client first connects
	welcome := protocol.Message{
		Type:      protocol.MessageTypeJoin,
		Timestamp: time.Now(),
		From:      "Server",
		To:        "",
		Text:      "Connected successfully. Welcome!",
	}
	if err := conn.SendMessage(welcome); err != nil {
		log.Println("Failed to send welcome message: ", err)
		return
	}

	for {
		rcv_msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Connection closed: ", err)
			return
		}
		fmt.Println("received: ", rcv_msg)

		switch rcv_msg.Type {
		case protocol.MessageTypeJoin:
			if err := conn.SendMessage(protocol.Message{
				Type:      protocol.MessageTypeJoin,
				Timestamp: time.Now(),
				From:      "Server",
				To:        cmp.Or(rcv_msg.From, conn.conn.RemoteAddr().String()),
				Text:      "You have joined the chat.",
			}); err != nil {
				log.Println("Failed to send join message: ", err)
			}
		case protocol.MessageTypeChat:
			if err := conn.SendMessage(protocol.Message{
				Type:      protocol.MessageTypeChat,
				Timestamp: time.Now(),
				From:      "Server",
				To:        cmp.Or(rcv_msg.From, conn.conn.RemoteAddr().String()),
				Text:      "You said: " + rcv_msg.Text,
			}); err != nil {
				log.Println("Failed to send text message: ", err)
			}
		case protocol.MessageTypeDisconnect:
			if err := conn.SendMessage(protocol.Message{
				Type:      protocol.MessageTypeDisconnect,
				Timestamp: time.Now(),
				From:      "Server",
				To:        cmp.Or(rcv_msg.From, conn.conn.RemoteAddr().String()),
				Text:      "You have left the chat.",
			}); err != nil {
				log.Println("Failed to send leave message: ", err)
			}
		}
	}
}
