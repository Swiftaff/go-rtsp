package main

import (
	"fmt"

	"github.com/swiftaff/hello/client"
	"github.com/swiftaff/hello/server"
)

func main() {
	url := "192.168.1.11:88"

	var which int
	fmt.Printf("0 = Server. 1 = Client\n")
	fmt.Scan(&which)

	switch which {
	case 0:
		server.Server(url)
	case 1:
		client.Client(url)

	}

}
