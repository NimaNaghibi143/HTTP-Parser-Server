package response

import (
	"fmt"
	"io"

	"http.nima.strive/internal/headers"
	"http.nima.strive/internal/request"
)

type Response struct {
}

type StatusCode int

const (
	StatusOk                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

// this handlererror is used to return the proper error message and status code
type HandlerError struct {
	StatusCode StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()

	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

func WriteHeader(w io.Writer, h *headers.Headers) error {
	b := []byte{}

	h.ForEach(func(n, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", n, v)
	})

	b = fmt.Append(b, "\r\n")
	_, err := w.Write(b)

	return err
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusLine := []byte{}

	switch statusCode {
	case StatusOk:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case StatusBadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Reqeust\r\n")
	case StatusInternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Errorq\r\n")
	default:
		return fmt.Errorf("Unrecognized error code!")
	}

	_, err := w.Write(statusLine)
	return err

}
