package client

import (
	"fmt"
	"net"
)

const (
	connType = "tcp"
)

//Client - basic tcp client
func Client(url string) {
	fmt.Printf("CLIENT: Opening connection to %s...\n", url)

	addr, err := net.ResolveTCPAddr("tcp", url)
	if err != nil {
		fmt.Printf("ResolveTCPAddr Error: %s\n", err.Error())
	}

	c, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		fmt.Printf("DialTCP Error: %s\n", err.Error())
	}
	fmt.Printf("Connected to %s\n", url)

	command := "Hello"
	writeBytes, err := c.Write([]byte(command))
	if err != nil {
		fmt.Println("Write Error: ", err, writeBytes)
	}
	fmt.Printf("Sent command: %s\n", command)

	fmt.Printf("...Closing connection to %s\n", url)
	c.Close()
}
