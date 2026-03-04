defaut:
    @just --list

# Run server
run:
    go run ./cmd/tcplistener/ | tee tcp.txt

# Send data
send:
    cat messages.txt | nc localhost 42069

# Send a get request to the server
send-get:
    curl http://localhost:42069/coffee