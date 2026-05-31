package server

import (
	"log"
	"net"
)

// the function to import into the main to create the Server
// this function returns a pointer to the server struct type
// adding the add string as the address to the addr field
func New(ip, port string) *Server {
	return &Server{
		ip:   ip,
		port: port,
	}
}

// Start function that actually powers on the server to listen
// and then accept the connections
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.ip+":"+s.port)
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
		newConn := NewConn(conn)
		go HandleConnection(newConn)
	}
}