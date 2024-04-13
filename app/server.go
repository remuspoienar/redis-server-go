package main

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/handler"
	"github.com/codecrafters-io/redis-starter-go/app/instances"
	. "github.com/codecrafters-io/redis-starter-go/app/internal"
	"net"
	"os"
)

func main() {
	instance := instances.New()
	props := instance.Props()

	address := fmt.Sprintf("0.0.0.0:%d", props.Port())
	l, err := net.Listen("tcp", address)

	if err != nil {
		fmt.Println("Failed to bind port on", address)
		os.Exit(1)
	}
	defer CloseConnections(l)

	if props.IsMaster() {
		fmt.Printf("[%s]Server is listening on %s\n", props.Role(), address)
	} else {
		conn := instance.ConnectToMaster()
		go handler.HandleConnection(conn, instance)
		fmt.Printf("[%s]Server is listening on %s\nas a replica for master %s\n", props.Role(), address, props.MasterAddress())
	}

	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}
		go handler.HandleConnection(conn, instance)
	}
}
