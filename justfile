defaut:
    @just --list

# Run server for GET request assignment
run-get:
    go run ./cmd/tcplistener/ | tee rawget.txt

# Send data
send:
    cat messages.txt | nc localhost 42069

# Send a GET request to the server
get:
    curl http://localhost:42069/coffee

# Run server for POST request assignment
run-post:
    go run ./cmd/tcplistener/ | tee rawpost.http

# Send a POST request to the server
post:
     curl -X POST -H "Content-Type: application/json" -d '{"flavor":"dark mode"}' http://localhost:42069/coffee

