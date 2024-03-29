package main

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"io"
	"net"
	"os"
	"strings"
	// Uncomment this block to pass the first stage
	// "net"
	// "os"
)

const PORT = "6379"
const PING = "PING"
const PONG = "PONG"
const DOCS = "COMMAND DOCS"
const ECHO = "ECHO"

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:"+PORT)

	if err != nil {
		fmt.Println("Failed to bind to port " + PORT)
		os.Exit(1)
	}
	defer l.Close()

	fmt.Println("Server is listening on port " + PORT)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		buf := make([]byte, 1024)

		n, err := conn.Read(buf)

		if err != nil {
			fmt.Println("Error reading data", err.Error())
			if err == io.EOF {
				return
			}
			continue
		}

		data := string(buf[:n])

		commandParts := resp.ParseCommand(data)
		command := strings.Join(commandParts, " ")

		fmt.Printf("parsed command: `%s`\n", command)

		if isCommand(command, PING) {
			conn.Write([]byte(resp.SimpleString(PONG)))
		} else if isCommand(command, DOCS) {
			conn.Write([]byte(resp.SimpleString("OK")))
		} else if isCommand(command, ECHO) {
			value := strings.Join(commandParts[1:], " ")
			conn.Write([]byte(resp.BulkString(value)))
		} else {
			conn.Write([]byte(resp.SimpleError("unknown command")))
		}
	}
}

func isCommand(input string, value string) bool {
	return strings.HasPrefix(strings.ToUpper(input), value)
}
