package response

import (
	"fmt"
	"io"

	"http.nima.strive/internal/headers"
)

type Response struct {
}

type Writer struct {
	writer io.Writer
}

type StatusCode int

const (
	StatusOk                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()

	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

// legacy header writer
// func WriteHeaders(w io.Writer, h *headers.Headers) error {
// 	b := []byte{}

// 	h.ForEach(func(n, v string) {
// 		b = fmt.Appendf(b, "%s: %s\r\n", n, v)
// 	})

// 	b = fmt.Append(b, "\r\n")
// 	_, err := w.Write(b)

// 	return err
// }

// legacy statusLine writer
// func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
// 	statusLine := []byte{}

// 	switch statusCode {
// 	case StatusOk:
// 		statusLine = []byte("HTTP/1.1 200 OK\r\n")
// 	case StatusBadRequest:
// 		statusLine = []byte("HTTP/1.1 400 Bad Reqeust\r\n")
// 	case StatusInternalServerError:
// 		statusLine = []byte("HTTP/1.1 500 Internal Server Errorq\r\n")
// 	default:
// 		return fmt.Errorf("Unrecognized error code!")
// 	}

// 	_, err := w.Write(statusLine)
// 	return err

// }

// Implementing our own writers and more flexible way of handling headers and status and body.

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
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

	_, err := w.writer.Write(statusLine)
	return err
}

func (w *Writer) WrtieHeader(h headers.Headers) error {
	b := []byte{}

	h.ForEach(func(n, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", n, v)
	})

	b = fmt.Append(b, "\r\n")
	_, err := w.writer.Write(b)

	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)

	return n, err
}
