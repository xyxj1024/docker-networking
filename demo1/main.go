package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

// Program entry point
func main() {
	// Obtain the port via program argument
	port := fmt.Sprintf(":%s", os.Args[1])

	// Create a TCP listener on the given port
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Failed to create listener, err : ", err)
		os.Exit(1)
	}
	fmt.Printf("Listening on %s\n", listener.Addr())

	// Listen for new connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection, err : ", err)
			continue
		}
		// Pass an accepted connection to a handler goroutine
		go connHandler(conn)
	}
}

// Handle the lifetime of a connection
func connHandler(conn net.Conn) {
	// Defer connection close until the handler returns
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		// Read client request data
		bytes, err := reader.ReadBytes(byte('\n'))
		if err != nil {
			if err != io.EOF {
				fmt.Println("Failed to read data, err : ", err)
			}
			return
		}
		fmt.Printf("Request: %s", bytes)

		// Response
		line := fmt.Sprintf("%s", bytes)
		fmt.Printf("Response: %s", line)
		conn.Write([]byte(line))
	}
}
