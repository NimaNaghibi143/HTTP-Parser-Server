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

## Phase 15

### Add to parse

it's time to add this header parser to our little state machine and parse the full set of headers.
ok now it's time for me to add this into the tcplistener and parse the live request's headers.

what we expect:

```bash
Request line:
- Method: METHOD
- Target: TARGET
- Version: VERSION
Headers:
- KEY: VALUE
- KEY: VALUE
...
```

We need to iterate the keys.

## Phase 16

### Parsing the body

in an HTTP-message:

```bash
HTTP-message = start-line CRLF
               *( field-line CRLF)
               CRLF
               [ message-body ]
```

our current code parses the:

* start-line
* field-line
* the extra CRLF between headers and the body

now it's time to parse the [ message-body ].

According to RFC9110 8.6:

> A user agent SHOULD send Content-Length in a request...

and should has a specific meaning in RFC per RFC2119:

All this to say, for our implementation we are going to **assume that if there is no *Content-length* header, there is no body present.** this is a safe assumption for our purposes, though might not be true in all cases in the wild.

## Phase 17

Now it's time for the end to end test for our full HTTP request parsing.

what we expect:

```bash
Request line:
- Method: METHOD
- Target: TARGET
- Version: VERSION
Headers:
- KEY: VALUE
- KEY: VALUE
Body:
BODY_STRING

```

## Some Notes

HTTP is the primary protocol on the web, and it's on top of TCP (http 1.1 & 2). but http 3 is actually on top of UDP which actually implements a form of TCP(quic).

> [!NOTE]
> Remember that HTTP responses follow the same HTTP message format. the only differece is that the start-line instead of a request-line.

> [!IMPORTANT]
> **A client SHOULD ignore the reason-phrase content because it is not a reliable channel for inforamtion** (it might be translated for a given locale, overwritten by intermidiaries, or discarded when the message is forwarded via other versions of HTTP). A server **MUST** send the space that separates the status-code from the reason-phrase even when the reason-phrase is absent.

## Refactorting

there are two points that we need to refactor:

1. Error are always handled as plain text responses
2. Headers are always the same

we need to improve our handler function to be more flexible for custom headers.

## This is not the best way of implementing the server

but anyway this is it for now! but we can use sub-handlers for based that we handle each request with a sub-handler based on the request header line and request target*. and write the proper status code for the request.

what i mean by the request line and request target:

```go
handler = subHandler.find(request.RequestLine.RequestTarget)
writer = bytes.NewBuffer()
handler(writer, headers, request)
```

yes there are going to be lots of sub handlers but what i think is:
when you are playing with this server stuff of course you are going to use the sub-handlers.
you are also going to want use sth like:

```go
recover Function func() any
```

when your server panics you would want your handlers to be caught in the recovery and by catching them in the recovery we are going to make sure the your server does not go down based on a panic but instead reports the panic.

## Chunked encoding

Let's pretend we have a server that recieves a request and that request wants to download a file. but the file you need to download is 10GB big. that's a big file! you can't just simply load that all in memory, be able to calculate the checksum and be able to send that our easily because if you remember we need sth named: **Content-Length** you need to know exactly how many bytes are coming down! but what if you don't even know long this file is ?! what happens of you are recieving this from another server or a different protocol that does not even specify the length you need a way to be able to send down data in such a way that you don't hav to specify the length up front, this is called ***chunked encoding***.

### Chunked Trailer Section

A trailer section allows the sender to include additional fields at the end of a chunked message in order to supply metadata that might be dynamically generted while the content is sent, such as a message integrity check, digital signature, or post-processig status.

A recipient that removes the chunked coding form a message **MAY** selectively retain or discard the recieved trailer fields. A recipient that retains a recieved trailer field **MUST** either store/forward the the trailer field separately from the recieved header fields or merge the recieved trailer field into the header section. A recipient **MUST NOT** merge a recieved trailer field into the header section unless it's corresponding header field definition explicitly permits and instructs how the trailer field value can be safely merged.

in the RFC it's refrered by ***chunked transfer coding***:

The chunked transfer coding wraps content in order to transfer it as a series of chunks, each with it's own size indicator, followed by an **OPTIONAL** of trailer section containing trailer fields. Chunked enables content streams of unknown size to be transferred as a sequence of length-delimited buffers which enables the sender to retain connection persistence and the recipient to know when it has recieved the entire message.

Each http body contains how many bytes are in the body and in the data bytes.

here is the format:

```bash
HTTP/1.1 200 ok
Content-Type: text/plain
Transfer-Encoding: chunked

<n>\r\n
<data of length n> n>\r\n
<n>\r\n
<data of length n> n>\r\n
... repeat ...

0\r\n
\r\n
```

Where `n` is just a hexadecimal number indicating the size of the chunk in bytes and data of length `n` is the actual data for that chunk. That pattern can be repeated as many times as necessary to send the entire message body.

Chunked encoding is most often used for:

* Streaming large amounts of data (like big files)
* Real-time updates (like a chat-style application)
* Sending data of unknown size (like a live feed)

### Testing

> [!NOTE]
> `https://httpbin.org/stream/x` streams x JSON responses to our server, making it a great way for us to test our chunked response implementation.

> [!NOTE]
> How to disable http2

```bash
curl --http1.1 -vvv https://httpbin.org/stream/5
```

### Testing the httpserver

```bash
echo -e "GET /httpbin/stream/100 HTTP/1.1\r\nHost: localhost:42069\r\nConnection: close\r\n\r\n" | nc localhost 42069
```

### Trailers implementation

You can have additional headers at the end of chunked encoding called trailers. They work the same way that headers do with one catch: you have to specidfy the names of the trailers in a Trailer header.

Trailers are often used to send information about the message body that **can't be known until the message body is fully sent.** For example, the hash of the message body.

## What is a Trailer attack?

### Context: HTTP 1.1 and Transfer Coding

HTTP/1.1 introduced persistent connections (keep-alive), allowing multiple requests/responses over a single TCP connection. To handle potentially unknown-length messages, two main mechanisms were introduced:

1.  **Content-Length:** Header specifying the exact number of bytes in the message body (and implicitly the entire message, including headers).
2.  **Chunked Transfer Coding:** A mechanism where the message body is divided into chunks of arbitrary length, each preceded by its size (in hex) and terminated by a zero-sized chunk. This allows streaming data.

**The Problem: Ambiguity with `Transfer-Encoding` and `Content-Length`**

>[!NOTE]
>When both `Transfer-Encoding: chunked` and `Content-Length` headers are present, their relationship is ambiguous according to the HTTP specification. The specification dictates that `chunked` is the *preferred* encoding if both are present, and `Content-Length` should be ignored. However, misconfigured servers or clients might not adhere strictly to this.

**What is a Trailer Attack?**

A Trailer Attack specifically exploits how certain servers handle the `Trailer` header in conjunction with chunked encoding, especially when misconfigured alongside `Content-Length`.

**Key Concepts:**

*   **Trailers:** HTTP allows adding header fields after the final chunk of a chunked transfer encoding. These are called "trailing headers". This is done using the `Trailer` header field to specify which headers should be treated as trailers.
    *   Example `Trailer` header: `Trailer: Content-Length`
    *   This means that the `Content-Length` header should be placed after the final chunk, not before the body.

**How the Trailer Attack Works (HTTP Request Smuggling)**

The attack leverages a server's incorrect handling of the `Trailer` header when receiving chunked-encoded requests, especially when a `Content-Length` header is also present (which should normally be ignored if `chunked` is present).

**Scenario:**

1. **Malicious Client:**
    *   The client sends an HTTP request using chunked transfer encoding.
    *   This request *also* includes a `Content-Length` header (which violates the standard if `chunked` is present, but attackers do this to confuse servers).
    *   Crucially, the client *also* includes a `Trailer` header specifying a header field that the targeted server might pay attention to (e.g., `Trailer: Content-Length`).

    Example Malicious Request (Client Perspective):

    ```bash
    POST /vulnerable HTTP/1.1
    Host: example.com
    Content-Length: 100   <--- Maliciously included, ignored by standard but present
    Transfer-Encoding: chunked
    Trailer: Content-Length <--- Specifies that 'Content-Length' is a trailer

    ... [chunked body data] ...
    ```

2.  **Server Processing (Exploiting Misconfiguration):**
    *   **Server A (Standard Compliant):** Sees `Transfer-Encoding: chunked`, ignores `Content-Length`. Processes the trailers correctly (placing `Content-Length` after the final chunk). This server is not vulnerable to this specific attack vector.
    *   **Server B (Misconfigured/Non-Compliant):**
        *   Sees the `Content-Length: 100` header first. It might *incorrectly* start expecting the request body to be exactly 100 bytes *before* even processing the `Transfer-Encoding: chunked` header.
        *   It then receives the chunked data. When it reaches the final zero-chunk, it looks at the `Content-Length` header it saw *earlier* and interprets the following data (which should be the trailer `Content-Length: ...`) as part of the body *after* the chunked data, because its initial `Content-Length` assumption is still active.

    Let's visualize the data stream (Client -> Server B):

    ```bash
    [Headers: ..., Content-Length: 100, Transfer-Encoding: chunked, Trailer: Content-Length, ...]
    [Chunk 1: 'a'*50]           <-- Server B sees this, thinks "Body so far: 50 bytes"
    [Chunk 2: 'a'*50]           <-- Server B sees this, thinks "Body so far: 100 bytes" (hits the declared Content-Length)
    [Final Chunk: 0\r\n\r\n]    <-- Server B: "Body received! Now process the trailer header ('Content-Length') as if it were part of the body stream?"

    [Trailer Header: Content-Length: 50] <--- This line is actually *after* the final chunk, but Server B thinks it's part of the body because it already hit the Content-Length
    [Chunk 3: 'a'*50]           <-- Server B sees this *after* processing the header it thought was part of the body, thinking it's more body data
    [Final Chunk: 0\r\n\r\n]    <-- Server B: "Body finished. Now process the response..."

    ```

    **Exploit Outcome:**

    *   Server B now believes the request body is the data from `Chunk 3` (50 bytes), not the initial chunks.
    *   Crucially, it uses the `Content-Length: 50` *from the trailer* as the perceived length of the request body.
    *   This allows the attacker to construct two *different* requests within one chunked transfer stream.

**Attack Vectors (Examples):**

1.  **Blind SQL Injection:** The attacker can structure the request so that part of the body (which the server misinterprets due to the smuggling) contains SQL commands designed to query the database.
2.  **Server Side-Channel Attacks:** The attacker crafts the request so that the body seen by one server (Server B) is different from the body seen by another server (Server A), potentially bypassing security restrictions or authentication mechanisms.
3.  **Access Control Bypass:** By smuggling credentials or other data into the request body that would be ignored or misinterpreted by standard-compliant servers, the attacker can bypass authentication or access controls.
4.  **Cross-Site Scripting (XSS):** Crafting the smuggled data to inject malicious scripts.

**Mitigations:**

1.  **Strict Chunked Transfer Encoding Parsing (Server Configuration):** This is the most critical mitigation.
    *   Servers must strictly adhere to the specification: `chunked` overrides `Content-Length`, and `Content-Length` should be ignored if `chunked` is present.
    *   Servers must correctly parse trailers, placing them *after* the final chunk, not interpreting them as part of the body.
    *   Examples:
        *   **Apache:** Configure `chunked_transfer_encoding` directive appropriately (though full compliance can be tricky). Use `Trailer` header correctly.
        *   **Nginx:** Use the `chunked_length` parameter in `http{...}` or `server{...}` blocks, and ensure `chunked` is prioritized over `Content-Length`. Validate `Trailer` headers properly.
        *   **IIS:** Ensure settings prioritize `chunked` over `Content-Length` and handle trailers correctly.

2.  **Remove/Ignore Malicious `Content-Length` Headers:** While not strictly compliant, some servers might choose to ignore `Content-Length` headers if `Transfer-Encoding: chunked` is present, preventing the ambiguity.

3.  **Client-Side Protections (Less Common):** Generally, fixing server-side compliance is preferred. Client-side checks are difficult because they rely on knowing the server's behavior.

4.  **Security Headers (Indirect):**
    *   `Expect-CT` (Connection Termination): Helps prevent man-in-the-middle attacks, but doesn't directly prevent smuggling.
    *   `Strict-Transport-Security` (HSTS): Ensures HTTPS, which uses HTTP/1.1, but doesn't fix server parsing issues.

5.  **Web Application Firewalls (WAFs):**
    *   Modern WAFs can detect patterns indicative of request smuggling, including Trailer attacks, by analyzing the differences between how clients and servers parse the request stream.