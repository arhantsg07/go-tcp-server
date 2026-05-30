package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type Server struct {
	addr string
}

// the function to import into the main to create the Server
// this function returns a pointer to the server struct type
// adding the add string as the address to the addr field
func New(addr string) *Server {
	return &Server{addr: addr}
}

// Start function that actually powers on the server to listen
// and then accept the connections
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	// we defer the close function, it executes only when the neighbouring
	// functions return
	defer listener.Close()

	// Now we create a while loop to handle multiple connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err)
			continue
		} else {
			log.Printf("Client connected: %v\n", conn.RemoteAddr())
		}
		go HandleConnection(conn)
	}
}

func HandleConnection(conn net.Conn) {

	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		n, err := conn.Read(buffer)

		if err != nil {
			if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
				fmt.Println("client timed out: ", nErr)
				return
			}

			if err == io.EOF {
				fmt.Println("client disconnected")
				return
			}

			fmt.Println("read error: ", err)
			return
		}
		fmt.Println("received: ", string(buffer[:n]))

		conn.SetWriteDeadline(time.Now().Add(30 * time.Second))

		msg := "hello this is Server"
		_, err = conn.Write([]byte(msg))
		if err != nil {
			fmt.Println("Write error: ", err)
			return
		}
	}
}
