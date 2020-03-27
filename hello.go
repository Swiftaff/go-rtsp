package main

import (
	"fmt"

	"github.com/swiftaff/hello/client"
	"github.com/swiftaff/hello/server"
)

func main() {
	url := "127.0.0.1:5000"

	var which int
	fmt.Println("0 = Server. 1 = Client")
	fmt.Scan(&which)

	switch which {
	case 0:
		server.Server(url)
	case 1:
		client.Client(url)

	}

}
