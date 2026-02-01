# HTTP Parser & Server

## Project Overview

This project implements a rudimentary HTTP parser and server built from scratch using the Go programming language. The primary objective is to gain a deep technical understanding of the HTTP protocol by bypassing the standard `net/http` package.

Instead of high-level abstractions, this implementation handles the protocol manually over raw TCP sockets using Go's `net` package. The core logic revolves around reading connections byte-by-byte in real-time to reconstruct and parse messages.

## Technical Architecture 

* **Language:** Go
* **Networking:** `net` (TCP/IP)
* **Concurrency Model:** Goroutines and Channels for non-blocking I/O processing.
* **Data Flow:** Stream-based processing (reading small chunks to construct lines).

---

## Development Phases

### Phase 1: Basic File I/O & Chunking

The initial implementation focuses on reading raw bytes from a local file source (`message.txt`) to establish the reading loop.

* Create `message.txt` populated with arbitrary text data.
* Implement a reader that consumes the file **8 bytes at a time**.
* Output the raw data to `stdout` in 8-byte chunks to visualize the stream.

### Phase 2: Line Buffering & State Management

Since network communication rarely aligns perfectly with fixed-byte chunks, this phase introduces state persistence to handle data continuity.

* Maintain a string buffer that persists across loop iterations.
* Upon reading an 8-byte chunk, split the data by newline characters (`\n`).
* Accumulate "partial" lines in the buffer until a full line is resolved.
* Format output line-by-line rather than chunk-by-chunk.

### Phase 3: Concurrency & Stream Abstraction

Refactor the reading logic into a reusable function designed for TCP streams. This function adopts a producer-consumer pattern using Go channels.

**Function Specification:**

> "I’ll give you a stream of lines, you just consume them."

1. **Input:** Accepts a TCP connection (or reader interface).
2. **Output:** Returns a `<-chan string` for immediate use by the caller.
3. **Behavior:**
* Spawns a **goroutine** to handle the read loop.
* Sends parsed lines to the channel one at a time.
* **Formatting:** Does *not* prefix lines with debug text (e.g., "read:") and strips trailing newlines.
* **Cleanup:** Closes the source file/connection when reading is complete. Closes the channel upon EOF to signal the main routine to exit.



### Phase 4: TCP Implementation

Transition from local file I/O to network I/O. While HTTP is the goal, it relies on TCP to guarantee packet ordering and delivery. This phase utilizes `net.Listen` to handle the transport layer.

**Server Logic:**

* **Port:** Listens on `:42069`.
* **Connection Handling:**
1. Wait to `.Accept()` a new connection.
2. Log "Connection Accepted" to the console.
3. Pass the connection to the channel-based reader (from Phase 3).
4. Print received lines to the console (raw output, terminating newline only).
5. Log "Connection Closed" when the channel closes.

Note on Stream Abstraction

Instead of reading 8 bytes at a time from a file, we are now reading from a connection. The underlying principle is the same: both represent a continuous stream of binary data.

While a file allows you to pull data (determining exactly when to read), a connection pushes data toward you. However, the interface is identical. The main function simply passes the connection object, and the logic processes the binary stream without modification.
 
* **Lifecycle:** The server runs in an infinite loop until terminated (SIGINT/`Ctrl+C`).
* **Resource Management:** Ensures the listener is `.Close()`'d on exit.

## Usage

To run the server and capture the output to a log file:

open 2 terminals:

Terminal 1:

```bash
go run main.go | tee tcp.txt

```

Terminal 2:

```bash
cat messages.txt | nc localhost 42069

````

