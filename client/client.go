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
		command = getCommand(input, url)
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
	//count, err := bufio.NewWriter(c).WriteString(command)
	count, err := c.Write([]byte(command))
	if err != nil {
		fmt.Println("Write Error: ", err, count)
	}
	fmt.Printf("Sent command %d chars\n\r%s\n", count, command)
}

func connectionRead(c *net.TCPConn) {
	//message := ""
	//for len(message) == 0 {
	for {
		fmt.Printf("Receiving1\n")
		//test, _ := c.Read([]byte("C"))
		//fmt.Printf("Message Received: %d", test)
		buf := bufio.NewReader(c)
		message, _ := buf.ReadString(' ')
		fmt.Printf("Receiving2\n%s", message)
		//if err != nil {
		//	fmt.Println("Read Error: ", err)
		//}
		//fmt.Printf("Receiving3\n")
		//fmt.Printf("Message Received: %s", message)
	}
}

func getUserInput() int {
	input := 0
	fmt.Println("2 = options. 3 = CFNL. 4. end")
	fmt.Scan(&input)
	return input
}

func getCommand(input int, url string) string {
	command := ""
	switch input {
	case 2:
		command = fmt.Sprintf("OPTIONS rtsp://%s RTSP/1.0\r\nCSeq: 1", url) //\r\nCseq: 1\r\nRequire: implicit-play\r\n", url)
		command = "Hello"
	case 3:
		command = "\r\n"
	case 4:
		command = "end"
	}
	return command
}
