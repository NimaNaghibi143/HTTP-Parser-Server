package server

import (
	"fmt"
	"net"
)

type Server struct {
}

func runServer(s *Server, listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
	}

}

func Serve(port uint16) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%d", port))

	if err != nil {
		return nil, err
	}

	server := &Server{}
	go runServer(server, listener)

	return server, nil
}

func (s *Server) Close() error {
	return nil
}
