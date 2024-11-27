package ftp

import "github.com/KacperMalachowski/study/internet-protocols/ftp/server/pkg/config"

type Server struct {
	address        string
	port           int
	users          config.Users
	minPassivePort int
	maxPassivePort int
	allowAnonymous bool

	sessions map[string]*Session
}

func NewServer(address string, port int, users config.Users, minPassivePort, maxPassivePort int, allowAnonymous bool) *Server {
	return &Server{
		address:        address,
		port:           port,
		users:          users,
		minPassivePort: minPassivePort,
		maxPassivePort: maxPassivePort,
		allowAnonymous: allowAnonymous,
	}
}
