package main

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/instances"
	. "github.com/codecrafters-io/redis-starter-go/app/internal"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
	"io"
	"net"
	"os"
	"strings"
)

func infoCommand(parts []string) string {
	var subCommand string
	if len(parts) < 2 {
		subCommand = "replication"
	} else {
		subCommand = parts[1]
	}
	switch subCommand {
	default:
		return props.ReplicationInfo()
	}
}

func handleConnection(conn net.Conn, db storage.Db) {
	defer CloseConnections(conn)

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

		switch {
		case IsCommand(command, "PING"):
			WriteString(conn, resp.SimpleString("PONG"))
		case IsCommand(command, "DOCS"):
			WriteString(conn, resp.SimpleString("OK"))
		case IsCommand(command, "ECHO"):
			value := strings.Join(commandParts[1:], " ")
			WriteString(conn, resp.BulkString(value))
		case IsCommand(command, "GET"):
			value := db.Get(commandParts[1])
			WriteString(conn, resp.BulkString(value))
		case IsCommand(command, "SET"):
			px := ParsePX(command)
			db.Set(commandParts[1], commandParts[2], px)
			WriteString(conn, resp.SimpleString("OK"))
		case IsCommand(command, "INFO"):
			value := infoCommand(commandParts)
			WriteString(conn, resp.BulkString(value))
		case IsCommand(command, "REPLCONF"):
			WriteString(conn, resp.SimpleString("OK"))
		case IsCommand(command, "PSYNC"):
			value := fmt.Sprintf("FULLRESYNC %s %d", props.ReplId(), props.ReplOffset())
			WriteString(conn, resp.SimpleString(value))
		default:
			WriteString(conn, resp.SimpleError("unknown command"))
		}
	}
}

var props instances.Properties
var db storage.Db

func main() {
	instance := instances.New()
	props = instance.Props()
	db = storage.NewDb()

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
		instance.ConnectToMaster()
		fmt.Printf("[%s]Server is listening on %s\nas a replica for master %s\n", props.Role(), address, props.MasterAddress())
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}
		go handleConnection(conn, db)
	}
}
