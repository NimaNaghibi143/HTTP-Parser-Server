# Project Overview

This is my first Golang project! and i want to learn this language by implementing HTTP by my self based on the RFC doc 9110 (HTTP semantics)

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

## Phase 6

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

## Phase 7

### HTTP post

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

## phase 8

### Let's do some testing

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

## Phase 9

### Parsing the Request Line

By building on top of TCP, we already have code that handles plain-text data, now we just need to take that plain text
and turn it into structured data, ensuring that it follows the HTTP protocol.

for example give the:

```bash
GET /coffee HTTP/1.1
Host: localhost:42069
User-Agent: curl/8.7.1
Accept: */*
```

we want our HTTP parser to return a struct that looks like this:

```go
type Request strcut {
   RequestLine RequestLine 
   Headers     map[string]string
   BOdy        []byte
}
```

## The Request Line

start-line is called the request-line and has a specific format:

```bash
HTTP-version = HTTP-name "/" DIGIT "." DIGIT
HTTP-name = %s"HTTP"
request-line = method SP request-target SP HTTP-version
```

**Note:** SP means "single space"

**Note** There is a note we should remember when parsing the strings in the http realm.
   new line in the http is \r\n not \n!
   but this is not ture all the time inside of the RFC! if the first line of the request is
   separated by \n you must assume that all following lines are separated by \n.

## Phase 10

### Parsing a Stream

TCP (and by extention, HTTP) is a streaming protocol, which means we recieve data in chunks and should be able to parse it as it comes in.
so we need to manage the state of our parser to handle incomplete reads. The challenge is that it needs to be smart enough to know that it's not done yet and keep reading until it gets the full request.

our buffer size here is tiny.If you look at the tests, you'll also recall that we added some test cases where only 1 or 2 bytes are read at a time. We want to test at these small buffer sizers to ensure that our parser can handle the edge cases where even sth as small as the request line is split across multiple reads.

**NOTE:**There is a difference between parsing and reading.

This was confusing for me as well but it's important to understand the difference. When we read, all
we're doing is moving the data from the reader (which in the case of HTTP is a network connection, but it could be a file as well) into our program. When we parse, we are taking that data and interpreting it (moving it from a []byte to a RequestLine struct.) Once it's parsed, we can discard it from the buffer to save memory.

### We built a state machine

In this phase we have built the state machine. the combination of `RequestFromReader` and `Request.parse` functions creates our state machine. We keep track of several piecese of state:

* How much data we have read from the `io.Reader` into the buffer
* How much data we have parsed from the buffer
* The current `state` actually i myself don't like the term "state" i think it's a loaded term(initialized, done, etc).

## Phase 11

### Connect the parsing

Ok so far we have got all the tests passing, now it's time to actually use our parser in our tcplistener.

```bash
# Run the tcp-listener and redirect the parsed request output.
go run ./cmd/tcplistener/ | tee temp/requestline.txt

# in another shell, send this request to it:
curl http://localhost:42069/nima/naghibi
```

## Phase 12

### Headers

in the RFC they go by the name "field-line". Each field line consists of a case-insensitive field name followed by a colon (":"), optional leading whitespace, the field line value, and optional trailing whitespace.

```bash
field-line = field-name ":" OWS field-value OWS
```

**NOTE** OWS means optional white space

**NOTE** according to the documentation there could be unlimited amount of whitespaces before and after the field-value (header-value).But when parsing a "field-name", **There must be no spaces between the ":" and the field-name**. so basically these are valid:

```bash
'Host: localhost:42069'
'           Host: localhost:42069'
```

but this is not:

```bash
'Host : localhost:42069 '
```

**NOTE:**I'm not a pro in terms of developing in GO but i think if we consider the HEADERS a seprate pacakge we are going to be happy because we are going to parse the headers both in the requests and responses.

## Phase 13

### Constraits

we need to implement **Case Insensivity**! the keys (not necessariy values) are case insensitive! if you use the hash map directly, you'll have to account for **Content-Length** and **content-length** being the same on your own.

### Valid chars

based on the RFC doc, field-name has an implicit definition of a token.
in other words, a field-name must contain only:

* Uppercase letters: A-Z
* Lowercase letters: a-z
* Digits: 0-9
* Special Chars: `!`, `$`, `#`, `%`, `^`, `&`, `*`, `+`, `-`, `_`, ...

and at least a length of 1.

## Phase 14

### Multiple values

ok for more conprehensive inforamtion i'mma use the RFC documentation it self for explaining this part.

### Field Lines and Combined Field Value

When a field name is only present once in a section, the combined "field value" for that field
consists of the corresponding field line value. When a field name is repeated within a section, its
combined field value consists of the list of corresponding field line values within that section,
concatenated in order, with each field line value separated by a comma.

```bash
# For example, this section:
Example-Field: Foo, Bar
Example-Field: Baz
```

contains two field lines, both with the field name "Example-Field". The first field line has a field
line value of "Foo, Bar", while the second field line value is "Baz". The field value for "Example-
Field" is the list "Foo, Bar, Baz".

**NOTE:** A server MUST NOT apply a request to the target resource until it receives the entire request
header section, since later header field lines might include conditionals, authentication
credentials, or deliberately misleading duplicate header fields that could impact request
processing.
