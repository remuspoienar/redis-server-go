package main

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
	"io"
	"net"
	"os"
	"strings"
)

const (
	PORT = "6379"
	PING = "PING"
	PONG = "PONG"
	DOCS = "COMMAND DOCS"
	ECHO = "ECHO"
	GET  = "GET"
	SET  = "SET"
)

func closeConnections(closable any) {
	switch s := closable.(type) {
	case net.Listener:
	case net.Conn:
		err := s.Close()
		if err != nil {
			fmt.Println("Error closing client")
		}
	default:
		fmt.Println("Nothing to close")
	}
}

func main() {
	db := storage.NewDb()
	l, err := net.Listen("tcp", "0.0.0.0:"+PORT)

	if err != nil {
		fmt.Println("Failed to bind to port " + PORT)
		os.Exit(1)
	}
	defer closeConnections(l)

	fmt.Println("Server is listening on port " + PORT)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handleConnection(conn, db)
	}
}

func handleConnection(conn net.Conn, db storage.Db) {
	defer closeConnections(conn)

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
			respond(conn, resp.SimpleString(PONG))
		} else if isCommand(command, DOCS) {
			respond(conn, resp.SimpleString("OK"))
		} else if isCommand(command, ECHO) {
			value := strings.Join(commandParts[1:], " ")
			respond(conn, resp.BulkString(value))
		} else if isCommand(command, GET) {
			value := db.Get(commandParts[1])
			respond(conn, resp.BulkString(value))
		} else if isCommand(command, SET) {
			db.Set(commandParts[1], commandParts[2])
			respond(conn, resp.SimpleString("OK"))
		} else {
			respond(conn, resp.SimpleError("unknown command"))
		}
	}
}

func isCommand(input string, value string) bool {
	return strings.HasPrefix(strings.ToUpper(input), value)
}

func respond(conn net.Conn, msg string) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("Error writing data")
	}
}
