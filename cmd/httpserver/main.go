package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"http.nima.strive/internal/request"
	"http.nima.strive/internal/response"
	"http.nima.strive/internal/server"
)

const port = 42069

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

		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/stream") {
			target := req.RequestLine.RequestTarget

			res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])

			if err != nil {
				body = respond500()
				status = response.StatusInternalServerError
			} else {
				w.WriteStatusLine(response.StatusOk)

				h.Delete("Content-length")
				h.Set("transfer-encoding", "chunked")
				h.Replace("content-type", "text/plain")
				w.WriteHeaders(*h)
				w.WriteBody(body)

				for {
					data := make([]byte, 32)

					n, err := res.Body.Read(data)
					if err != nil {
						break
					}

					w.WriteBody([]byte(fmt.Sprintf("%x\r\n", n)))
					w.WriteBody(data[:n])
					w.WriteBody([]byte("\r\n"))
				}

				w.WriteBody([]byte("0\r\n\r\n"))
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
