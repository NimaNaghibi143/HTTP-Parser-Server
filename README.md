# HTTP Parser & Server (From Scratch in Go)

## Project Overview

This project implements a rudimentary **HTTP parser and TCP server** from scratch using the Go programming language. The primary goal is **deep, mechanical understanding** of how HTTP actually works by deliberately avoiding Go’s high-level `net/http` package.

Instead of abstractions, this project works directly on **raw TCP streams** using Go’s `net` package. Data is read incrementally (byte chunks), buffered, reconstructed into lines, and then consumed by higher-level logic — mirroring how real HTTP servers operate internally.

This is a learning-first project: correctness, clarity, and understanding take priority over features.

## Technical Architecture

* **Language:** Go
* **Networking:** `net` (TCP/IP)
* **Concurrency Model:** Goroutines + Channels
* **I/O Model:** Stream-based (incremental reads)
* **Parsing Strategy:** Stateful line buffering

## Design Philosophy

> Treat everything as a stream.

Files and network connections are fundamentally the same abstraction: a continuous stream of bytes. This project is built around that idea.

By reading from streams in small, fixed-size chunks, we:

* Simulate real network behavior
* Avoid assumptions about packet boundaries
* Learn why buffering and state matter

## Development Phases

## Phase 1: Basic File I/O & Chunking

Goal: understand raw byte streams and incremental reads.

### Steps

* Create a local file `message.txt` with arbitrary text
* Open the file using Go’s `os` package
* Read **8 bytes at a time** in a loop
* Print each chunk directly to `stdout`

### Key Insight

> Reads do not align with logical data boundaries.

You may receive partial words, partial lines, or multiple logical units in a single read.

## Phase 2: Line Buffering & State Management

Goal: reconstruct logical lines from arbitrary byte chunks.

### Approach

* Maintain a persistent string buffer
* Append newly-read bytes to the buffer
* Split the buffer on `\n`
* Emit complete lines
* Preserve partial lines for the next iteration

### Key Insight

> Stream processing requires state.

Without buffering, you cannot safely parse line-based protocols like HTTP.

## Phase 3: Concurrency & Stream Abstraction

Goal: decouple reading from consuming.

### Abstraction

> "I’ll give you a stream of lines — you just consume them."

### Function Behavior

* Accepts an `io.Reader` (file or TCP connection)
* Spawns a **goroutine** to handle reading
* Reads incrementally
* Emits parsed lines over a `<-chan string`
* Strips trailing newlines
* Closes the reader on EOF
* Closes the channel to signal completion

This turns blocking I/O into a clean, composable stream.

## Phase 4: TCP Server Implementation

Goal: move from file streams to real network streams.

### Server Configuration

* **Protocol:** TCP
* **Port:** `:42069`

### Server Lifecycle

1. Call `net.Listen("tcp", ":42069")`
2. Accept incoming connections in a loop
3. Log `Connection Accepted`
4. Pass the connection to the line-stream reader
5. Print received lines to `stdout`
6. When the channel closes:

   * Log `Connection Closed`
   * Clean up resources

The server runs indefinitely until interrupted (`Ctrl+C`).

## Phase 5: TCP Fundamentals

### What is TCP?

**TCP (Transmission Control Protocol)** is a **reliable, ordered, connection-oriented** transport-layer protocol. It ensures that data arrives:

* Completely
* In order
* Without duplication

TCP exposes data to applications as a **byte stream**, not discrete packets.

---

### Reliability & Sliding Window

TCP splits data into segments and uses a **sliding window** to control how many segments may be in flight.

Example:

* Total segments: 8
* Window size: 4

Process:

1. Send segments 1–4
2. Receive ACKs from the receiver
3. Slide the window forward
4. Send segments 5–8

If a segment is lost, it is retransmitted.

---

### Handshake & State

TCP requires a **three-way handshake**:

1. SYN
2. SYN-ACK
3. ACK

After this, both sides maintain connection state until closed.

---

### What is UDP?

**UDP (User Datagram Protocol)** is:

* Connectionless
* Stateless
* Unreliable
* Unordered

There are no ACKs, no retransmissions, and no guarantees.

Any reliability must be implemented by the application itself.

---

### TCP vs UDP

| Feature            | TCP    | UDP    |
| ------------------ | ------ | ------ |
| Connection         | Yes    | No     |
| Handshake          | Yes    | No     |
| Reliability        | Yes    | No     |
| In Order           | Yes    | No     |
| Flow Control       | Yes    | No     |
| Congestion Control | Yes    | No     |
| Raw Speed          | Slower | Faster |

---

### Why HTTP Uses TCP

HTTP depends on TCP because:

* Headers must arrive before bodies
* Requests must be complete
* Responses must be ordered

Your line-based stream reader works *because TCP guarantees a reliable byte stream*, even though the underlying packets may arrive fragmented.

---

## Usage

### Terminal 1 (Run Server)

```bash
go run ./cmd/tcplistener/ | tee tcp.txt
```

### Terminal 2 (Send Data)

```bash
cat messages.txt | nc localhost 42069
```


### Phase 6:

At the heart of HTTP is the HTTP-message: the format that the text in an HTTP request or response must use.

```bash
start-line CRLF 
*( field-line CRLF)
CRLF
[message body]
```

CRLF (written in plain text as \r\n -> like primeagan said "Registered Nurse") is a carriage return followed by a line feed, It's a Microsoft Windows (and HTTP) style newline character.

```bash
# in one shell do this:

go run ./cmd/tcplistener | tee /tmp/rawget.http

# in another shell:
 
curl http://localhost:42069/coffee
```

### Phase 7:

## HTTP post 

curl is a command line tool for making HTTP requests. if you cat the tcp.txt that we have just created, 
you should have sth like this:

```bash 
GET /goodies HTTP/1.1
Host: localhost:42069
User-Agent: curl/8.6.0
Accept: */*
```

This is what a raw HTTP message looks like - specifically a raw HTTP GET request.

Ok now let's implement a raw HTTP POST request.

1- First let's run our tcp listener!

```bash
go run ./cmd/tcplistener | tee /tmp/rawpost.http
```

2- send a post request:

```bash
 curl -X POST -H "Content-Type: application/json" -d '{"flavor":"dark mode"}' http://localhost:42069/coffee
```

### phase 8: 

## Let's do some testing! 

in our case that we are building a raw http server we need tests! so we know what we are implementing behaves the way we expect it to behave!
so for the sake of our own comfort and nervs we are going to avoid test tables! those nested if and elses and break points and unused prints!.
so we need our test to be as declarative as possible and we are going to avoid logic because logic is where the mistakes and error appear! so we are going to get rid of those.

1- let's create a internal directory and a request directory inside of it:

```bash
mkdir -p ./internal/request
```

2- Then create a request.go file. declare that it's part of the request package.
3- Create a request_test.go file, it's also part of the request package. and our test will go there.

4- install the "testify" package as a dependency in your module.

```bash
go get -u github.com/stretchr/testify/assert
```
