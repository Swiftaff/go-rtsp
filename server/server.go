package server

import (
	"fmt"
	"net"
	"os"
)

//Server - basic tcp server
//from https://coderwall.com/p/wohavg/creating-a-simple-tcp-server-in-go
func Server(domain string, port int) {
	domainport := fmt.Sprintf("%s:%d", domain, port)
	fmt.Printf("SERVER: Listening on %s...\n", domainport)

	// Listen for incoming connections.
	l, err := net.Listen("tcp", domainport)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + domainport)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	len := 1
	for len > 0 {
		buf := make([]byte, 1024)
		// Read the incoming connection into the buffer.
		reqLen, err := conn.Read(buf)
		len = reqLen
		if err != nil {
			fmt.Println("Error reading:", err.Error())
		}

		returnMessage := fmt.Sprintf("Message received of %d bytes: %s\n", reqLen, buf[0:reqLen])
		fmt.Printf(returnMessage)
		// Send a response back to person contacting us.
		conn.Write([]byte(returnMessage))
	}
	// Close the connection when you're done with it.
	conn.Close()
	os.Exit(1)
}
