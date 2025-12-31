package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/uygardeniz/http-from-tcp/internal/request"
	"github.com/uygardeniz/http-from-tcp/internal/response"
)

type Server struct {
	Port     int
	Listener net.Listener
	handler  Handler

	closed atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		Port:     port,
		Listener: listener,
		handler:  handler,
	}

	go server.listen()

	return server, nil

}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.Listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}

			log.Printf("Accept error: %v", err)
			continue
		}

		go func(conn net.Conn) {
			s.handle(conn)

		}(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}
		hErr.Write(conn)
		return
	}

	responseWriter := response.NewWriter(conn)
	hErr := s.handler(responseWriter, req)

	if hErr != nil {
		hErr.Write(conn)
		return
	}
}

type Handler func(w *response.Writer, req *request.Request) *HandlerError
type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (hErr *HandlerError) Write(writer io.Writer) {
	response.WriteStatusLine(writer, hErr.StatusCode)
	body := fmt.Appendf([]byte{}, "message: %s", hErr.Message)
	headers := response.GetDefaultHeaders(len(body))
	response.WriteHeaders(writer, headers)
	writer.Write(body)
}
