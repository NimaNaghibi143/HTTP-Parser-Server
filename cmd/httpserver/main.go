package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"http.nima.strive/internal/request"
	"http.nima.strive/internal/response"
	"http.nima.strive/internal/server"
)

const port = 42069

func main() {
	s, err := server.Serve(port, func(w io.Writer, req *request.Request) *server.HandlerError {
		// curl -v  http://localhost:42069/yourProblem
		if req.RequestLine.RequestTarget == "/yourProblem" {
			return &server.HandlerError{
				StatusCode: response.StatusBadRequest,
				Message:    "Ok that's your bad!\n",
			}
			// curl -v  http://localhost:42069/myProblem
		} else if req.RequestLine.RequestTarget == "/myProblem" {
			return &server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Message:    "Ok, i get it my bad\n",
			}
		} else {
			w.Write([]byte("ALl good!\n"))
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error starting the server: %v", err)
	}

	defer s.Close()
	log.Println("Server started on port:", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped!")

}
