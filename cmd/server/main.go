package main

import (
	"github.com/arhantsg07/go-tcp-server/internal/server"
)

func main() {

	srv := server.New("127.0.0.1:8080")
	srv.Start()
}
