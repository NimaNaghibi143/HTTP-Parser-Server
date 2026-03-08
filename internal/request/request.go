package request

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"unicode"

	"http.nima.strive/internal/headers"
)

type parserState string

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateDone    parserState = "done"
	StateError   parserState = "error"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	state       parserState
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
	}
}

var ErrorMalformedRequestLine = fmt.Errorf("Malformed request-line!")
var ErrorUnsupportedHttpVersion = fmt.Errorf("Unsupported Http version!")
var ErrorRequestInErrorState = fmt.Errorf("Request in error state!")

// registered nurse:)
var SEPARATOR = []byte("\r\n")

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	//how many bytes we have read.
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrorMalformedRequestLine
	}

	for _, c := range string(parts[0]) {
		if !unicode.IsUpper(c) {
			return nil, 0, ErrorMalformedRequestLine
		}
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ErrorUnsupportedHttpVersion
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil

}

// determine how much of the buffer we have parsed and error.
func (r *Request) parse(data []byte) (int, error) {
	read := 0
	// this is the state machine for parsing the http request.
outer:
	for {
		currentData := data[read:]
		slog.Info("Request#parse", "currentData", string(currentData))

		switch r.state {
		case StateError:
			return 0, ErrorRequestInErrorState

		case StateInit:
			slog.Info("Request#parse state-init", "read", 0)
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n
			r.state = StateHeaders

		case StateHeaders:
			slog.Info("Request#parse state-headers", "read", read)
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n

			if done {
				r.state = StateDone
			}

		case StateDone:
			break outer

		default:
			panic("I just fucked it up some where in the code!")
		}
	}

	slog.Info("Request#parse -- return", "state", r.state, "read", read)
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	//We want to simulate reading slowly over time! so we are going to use a for loop instead
	//of reading the whole message at once. because we don't know how long it's going to take!
	//becasue we may get stuck! we only need to read through the headers! the body parsing does not
	//need to happen right away. you can send off a message: hey i got a new request, it's to this
	//path, it's a POST for example it has these headers.
	//so after the first line you already know what handler to call, you have the path and the verb.

	// NOTE: buffer could get overrun... a header that exeeds 1k would do that ...
	// or the body.
	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}
		readN, err := request.parse(buf[:bufLen+n])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen+n])
		bufLen = bufLen + n - readN
	}

	return request, nil
}
