package main

import (
	"flag"
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	DefaultPort = 6379
	PING        = "PING"
	PONG        = "PONG"
	DOCS        = "COMMAND DOCS"
	ECHO        = "ECHO"
	GET         = "GET"
	SET         = "SET"
	INFO        = "INFO"
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

var role = "master"

func main() {
	var port int
	var masterHost string
	var masterAddress string
	flag.IntVar(&port, "port", DefaultPort, "Port to run server(positive integer)")
	flag.StringVar(&masterHost, "replicaof", "", "Provide master address to start in replica mode)")
	flag.Parse()

	db := storage.NewDb()
	address := fmt.Sprintf("0.0.0.0:%d", port)
	l, err := net.Listen("tcp", address)

	if err != nil {
		fmt.Println("Failed to bind port on " + address)
		os.Exit(1)
	}
	defer closeConnections(l)

	if masterHost != "" {
		var masterPort string

		if len(flag.Args()) > 0 {
			masterPort = flag.Args()[0]
		} else {
			fmt.Println("Invalid master address")
			os.Exit(1)
		}
		masterAddress = fmt.Sprintf("%s:%s", masterHost, masterPort)
		role = "slave"
		fmt.Printf("[%s]Server is listening on %s\nas a replica for master %s\n", role, address, masterAddress)
	} else {
		fmt.Printf("[%s]Server is listening on %s\n", role, address)
	}

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

		switch {
		case isCommand(command, PING):
			respond(conn, resp.SimpleString(PONG))
		case isCommand(command, DOCS):
			respond(conn, resp.SimpleString("OK"))
		case isCommand(command, ECHO):
			value := strings.Join(commandParts[1:], " ")
			respond(conn, resp.BulkString(value))
		case isCommand(command, GET):
			value := db.Get(commandParts[1])
			respond(conn, resp.BulkString(value))
		case isCommand(command, SET):
			px := parsePX(command)
			db.Set(commandParts[1], commandParts[2], px)
			respond(conn, resp.SimpleString("OK"))
		case isCommand(command, INFO):
			value := infoCommand(commandParts)
			respond(conn, resp.BulkString(value))
		default:
			respond(conn, resp.SimpleError("unknown command"))
		}
	}
}

func isCommand(input string, value string) bool {
	return strings.HasPrefix(strings.ToUpper(input), value)
}

func infoCommand(parts []string) string {
	var subCommand string
	if len(parts) < 2 {
		subCommand = "replication"
	} else {
		subCommand = parts[1]
	}
	switch subCommand {
	default:
		return fmt.Sprintf("role:%s", role)
	}
}

func respond(conn net.Conn, msg string) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("Error writing data")
	}
}

func parsePX(command string) int64 {
	var pxRaw string

	pxArgs := strings.Split(strings.ToUpper(command), "PX")
	if len(pxArgs) == 1 {
		pxRaw = "-1"
	} else {
		pxRaw = strings.TrimSpace(pxArgs[1])
	}
	px, err := strconv.ParseInt(pxRaw, 10, 32)
	if err != nil {
		fmt.Println("Could not parse PX from command: ", command)
		px = -1
	}
	return px
}
