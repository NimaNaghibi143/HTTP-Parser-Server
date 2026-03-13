package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"http.nima.strive/internal/headers"
	"http.nima.strive/internal/request"
	"http.nima.strive/internal/response"
	"http.nima.strive/internal/server"
)

const port = 42069

func toStr(bytes []byte) string {
	out := ""

	for _, b := range bytes {
		out += fmt.Sprintf("%02x", b)
	}

	return out
}

func respond400() []byte {
	return []byte(`<html>
    			<head>
        			<title>400 Bad Request</title>
    			</head>
    			<body>
        			<h1>Bad Request</h1>
        			<p>Your request kinda is fucked up.</p>
    			</body>   
			</html>`)
}

func respond500() []byte {
	return []byte(`			<html>
    			<head>
        			<title>500 internal server error</title>
    			</head>
    			<body>
        			<h1>Internal server error</h1>
        			<p>This time i just fucked it up! sorry:).</p>
    			</body>   
			</html>`)
}

func respond200() []byte {
	return []byte(`			<html>
    			<head>
        			<title>200 Bad Request</title>
    			</head>
    			<body>
        			<h1>Finally did it! </h1>
        			<p>Your request just nailed it ma boy!.</p>
    			</body>   
			</html>`)
}

func main() {
	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) {

		h := response.GetDefaultHeaders(0)
		body := respond200()
		status := response.StatusOk

		if req.RequestLine.RequestTarget == "/yourProblem" {

			body = respond400()
			status = response.StatusBadRequest

		} else if req.RequestLine.RequestTarget == "/myProblem" {

			body = respond500()
			status = response.StatusInternalServerError

			// /httpbin/stream for the previous phase of the trailers
		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
			target := req.RequestLine.RequestTarget

			res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])

			if err != nil {
				body = respond500()
				status = response.StatusInternalServerError
			} else {
				w.WriteStatusLine(response.StatusOk)

				//  The specification dictates that chunked is the preferred encoding if both are present, and Content-Length should be ignored.
				h.Delete("Content-length")
				h.Set("transfer-encoding", "chunked")
				h.Replace("content-type", "text/plain")
				h.Set("Trailer", "X-Content-SHA256")
				h.Set("Trailer", "X-Content-Length")
				w.WriteHeaders(*h)

				fullBody := []byte{}

				for {
					data := make([]byte, 32)

					n, err := res.Body.Read(data)
					if err != nil {
						break
					}

					fullBody = append(fullBody, data[:n]...)

					w.WriteBody([]byte(fmt.Sprintf("%x\r\n", n)))
					w.WriteBody(data[:n])
					w.WriteBody([]byte("\r\n"))
				}

				// agter running the server to test the trailes: curl -X GET http://localhost:42069/httpbin/html
				// remeber to check the prefix for the headers! (/httpbin/stream or /httpbin/html.)
				w.WriteBody([]byte("0\r\n"))
				trailers := headers.NewHeaders()
				out := sha256.Sum256(fullBody)
				trailers.Set("X-Content-SHA256", toStr(out[:]))
				trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
				w.WriteHeaders(*trailers)

				return
			}
		}

		h.Replace("Content-length", fmt.Sprintf("%d", len(body)))
		h.Replace("Content-type", "text/html")

		w.WriteStatusLine(status)
		w.WriteHeaders(*h)
		w.WriteBody(body)

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
