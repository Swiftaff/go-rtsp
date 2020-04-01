package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/swiftaff/hello/client"
	"github.com/swiftaff/hello/env"
	"github.com/swiftaff/hello/server"
)

func main() {
	env.Env() //secrets!
	domain := os.Getenv("domain")
	port, _ := strconv.Atoi(os.Getenv("port"))
	username := os.Getenv("username")
	password := os.Getenv("password")

	var which int
	fmt.Printf("0 = Server. 1 = ManualClient. 2 = AutoClient\n")
	fmt.Scan(&which)

	switch which {
	case 0:
		server.Server(domain, port)
	case 1:
		client.ManualClient(domain, port, username, password)
	case 2:
		client.Client(domain, port, username, password)
	}

}
