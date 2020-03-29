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
	scanner := bufio.NewScanner(bufio.NewReader(c))
	data := ""
	line := ""
	for scanner.Scan() {
		line = scanner.Text()
		data += line + "\n"
		if len(line) == 0 {
			break
		}
	}
	fmt.Printf("RECEIVED:\n%s", data)
}

func getUserInput() int {
	input := 0
	fmt.Println("1:OPTIONS 2:DESCRIBE 9:end")
	fmt.Scan(&input)
	return input
}

func getCommand(input int, url string) string {
	command := ""
	switch input {
	case 1:
		command = fmt.Sprintf("OPTIONS rtsp://%s/videoMain RTSP/1.0\r\nCSeq: 1\r\n\r\n", url)
		/*
			OPTIONS rtsp://127.0.0.1:5000/videoMain RTSP/1.0
			CSeq: 1
		*/
	case 2:
		command = fmt.Sprintf("DESCRIBE rtsp://%s/videoMain RTSP/1.0\r\nAccept: application/sdp\r\nCSeq: 2\r\n\r\n", url)
		/*
			DESCRIBE rtsp://192.168.1.11:88/videoMain RTSP/1.0
			Accept: application/sdp
			CSeq: 2
		*/
	case 3:
		command = fmt.Sprintf("DESCRIBE rtsp://%s/videoMain RTSP/1.0\r\nAccept: application/sdp\r\nCSeq: 3\r\n\r\n", url)

	case 9:
		command = "end"
	}
	return command
}

/*
DESCRIBE rtsp://192.168.1.11:88/videoMain RTSP/1.0
Accept: application/sdp
CSeq: 3
User-Agent: Lavf58.38.101
Authorization: Digest username="foscam", realm="Foscam IPCam Living Video", nonce="d4dea19512f55715153c33444c826135", uri="rtsp://192.168.1.11:88/videoMain", response="56b96e5d53070d39b54c82a2184501bc"
*/
