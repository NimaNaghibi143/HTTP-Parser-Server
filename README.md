# HTTP-Parser-Server
Introduction:
I’m about to start a fun project where I build an HTTP parser and server from scratch using the Go programming language. I won’t be using the net/http package, because I want to build everything myself. However, I will use the TCP library, since HTTP is built on top of TCP.

By the time I finish this project, I will have learned a lot about HTTP not only from a theoretical perspective, but also from a practical and technical standpoint.

so what is the project purpose? 
hmm basically our server is going read form a connection byte by byte in real time.

First Step:

1- create a message.txt file and then put a random text on it.
2- now we want our program to to read the messages.txt 8 bytes at a time and we are going to print the data back to stdout in 8 bytes chuncks.

- Reading 8 bytes at a time is a good start, but 8 byte chuncks are not how people tend to communicate. 

Second step:

ok let's update our code to continue to read 8 bytes at a time. but now let's format the output line by line.

1- Create a string varaible to hold the contents of the current line of the file, it needs to persist between reads (loop iteration)

2- After reading 8 bytes, split the data on newlines to create a slice of stirngs - we are going to call these split sections "parts". There will typically only be one or two "parts" because we are only reading 8 bytes at a time.

Third step:

ok let's refactor the code and create a reusable function that reads lines form a TCP connection.

it should contain all the logic i have already created and that keeps track of the current line's contents, reads 8 bytes at a time, etc. but!

Note: "I’ll give you a stream of lines, you just consume them."

- it should create a channel of strings 
- it does the reading loop inside a a gotroutine 
- it does not prefix the lines with the "read: " or add a newline character at the end.
  instead, it sends one line at a time to the channel.
- the goroutine exits when it reaches the end of the file, and the channel is closed.
- the function returns the channel for immediate use by the caller.
- the function closes the file when it's done reading (don't close it when the channel is returned.)
- 