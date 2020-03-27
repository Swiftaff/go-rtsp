package server

import (
	"fmt"
	"net"
	"os"
)

const (
	connType = "tcp"
)

//Server - basic tcp server
//from https://coderwall.com/p/wohavg/creating-a-simple-tcp-server-in-go
func Server(url string) {
	fmt.Printf("SERVER: Listening on %s...\n", url)

	// Listen for incoming connections.
	l, err := net.Listen(connType, url)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + url)
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
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	reqLen, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}

	message := fmt.Sprintf("Message received of %d bytes: %s%s", reqLen, buf, buf)
	// Send a response back to person contacting us.
	conn.Write([]byte(message))
	// Close the connection when you're done with it.
	//conn.Close()
}
