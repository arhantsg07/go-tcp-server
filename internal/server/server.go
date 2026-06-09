// This file contains the server definition struct

package server

import "sync"

type Server struct {
	ip     string
	port   string
	client map[string]*Connection
	mu     sync.Mutex
}
