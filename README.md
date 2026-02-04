## Phase 5: TCP Fundamentals

### What is TCP?

**TCP (Transmission Control Protocol)** is a **reliable, ordered, and connection-oriented** transport-layer protocol. Its job is to ensure that data sent from one machine arrives at another **intact**, **in order**, and **without duplication**.

When an application sends data over TCP, the data is split into smaller units called **segments** (often informally called packets). TCP manages how these segments are transmitted, acknowledged, retransmitted if lost, and reassembled on the receiving side.

---

### Reliability and Ordering

TCP guarantees:

* **Reliable delivery** – lost segments are detected and retransmitted
* **In-order delivery** – segments are reassembled in the correct sequence
* **Flow control** – the sender does not overwhelm the receiver
* **Congestion control** – the network is not overwhelmed

---

### Sliding Window

TCP uses a mechanism called a **sliding window** to control how much data can be "in flight" at any given time.

Example:

* You want to send **8 segments** total
* The sliding window size is **4**

Process:

1. The sender transmits segments 1–4
2. These segments are now "in flight"
3. The receiver sends **ACKs (acknowledgements)** for received segments
4. As ACKs arrive, the window "slides" forward
5. The sender transmits segments 5–8

If a segment is lost:

* The receiver does not ACK it
* The sender retransmits the missing segment

This mechanism is what makes TCP reliable.

---

### Handshake and Connection State

TCP is **stateful** and requires a **three-way handshake** before data transmission:

1. SYN – client requests a connection
2. SYN-ACK – server acknowledges and responds
3. ACK – client confirms

Only after this handshake is the connection established and data allowed to flow.

---

### What is UDP?

**UDP (User Datagram Protocol)** is a **connectionless** and **stateless** transport-layer protocol.

Key characteristics:

* No handshake
* No acknowledgements (ACKs)
* No retransmissions
* No ordering guarantees

The sender simply sends datagrams, and the receiver processes whatever arrives.

If data is lost, duplicated, or arrives out of order, UDP does **nothing** to correct it.

---

### Performance Trade-offs

UDP is often described as "faster" than TCP because:

* There is no handshake overhead
* There is no waiting for ACKs
* There is no retransmission logic

However, this means:

* The application must handle reliability if needed
* The receiver may see incomplete or corrupted data

Some systems build custom reliability on top of UDP using sequence numbers and **NACKs (negative acknowledgements)**, but this logic lives **above** UDP, not inside it.

---

### TCP vs UDP Comparison

| Feature            | TCP    | UDP    |
| ------------------ | ------ | ------ |
| Connection         | Yes    | No     |
| Handshake          | Yes    | No     |
| Reliability        | Yes    | No     |
| In-order           | Yes    | No     |
| Flow Control       | Yes    | No     |
| Congestion Control | Yes    | No     |
| Speed (Raw)        | Slower | Faster |

---

### Why This Matters for HTTP

HTTP relies on TCP because:

* Requests and responses must be complete
* Headers must arrive before bodies
* Data must be ordered and reliable

Your server implementation depends on TCP’s guarantees so that reading from a stream (byte-by-byte or line-by-line) behaves predictably, even though the underlying packets may arrive fragmented or delayed.

This is why your stream abstraction works: **TCP turns packets into a reliable byte stream**.

### Phase 5 :UDP Sender 

we are going to create a new program that yeets UDP packets at a server. We want to fully understand the difference between the TCP and UDP.

