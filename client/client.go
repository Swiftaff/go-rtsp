package client

import (
	"bufio"
	"fmt"
	"net"
)

const (
	connType = "tcp"
)

//Client - basic tcp client
func Client(url string) {
	c := connectionOpen(url)
	command := "Hello"
	for command != "end" {
		input := getUserInput()
		command = getCommand(input)
		connectionWrite(c, command)
		connectionRead(c)
	}
	connectionClose(c, url)
}

func connectionOpen(url string) *net.TCPConn {
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
	return c
}

func connectionClose(c *net.TCPConn, url string) {
	fmt.Printf("...Closing connection to %s\n", url)
	c.Close()
}

func connectionWrite(c *net.TCPConn, command string) {
	writeBytes, err := c.Write([]byte(command))
	if err != nil {
		fmt.Println("Write Error: ", err, writeBytes)
	}
	fmt.Printf("Sent command: %s\n", command)
}

func connectionRead(c *net.TCPConn) {
	message := ""
	for len(message) == 0 {
		message, _ = bufio.NewReader(c).ReadString('\n')
		fmt.Print("Message Received:", string(message))
	}
}

func getUserInput() int {
	input := 0
	fmt.Println("2 = options. 3 = world. 4. end")
	fmt.Scan(&input)
	return input
}

func getCommand(input int) string {
	command := ""
	switch input {
	case 2:
		command = "Hello"
	case 3:
		command = "world"
	case 4:
		command = "end"
	}
	return command
}
