package handler

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/instances"
	. "github.com/codecrafters-io/redis-starter-go/app/internal"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"io"
	"net"
	"strings"
)

func infoCommand(parts []string, props instances.Properties) string {
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

func HandleConnection(conn net.Conn, instance instances.Instance) {
	defer CloseConnections(conn)

	db := instance.Db()
	props := instance.Props()

	for {
		buf := make([]byte, 4096)

		n, err := conn.Read(buf)

		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Println("Error reading data", err.Error())
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

			if props.IsMaster() {
				WriteString(conn, resp.SimpleString("OK"))
				instance.PropagateCommand(buf[:n])
			}
		case IsCommand(command, "INFO"):
			value := infoCommand(commandParts, props)
			WriteString(conn, resp.BulkString(value))
		case IsCommand(command, "REPLCONF"):
			if props.IsMaster() {
				WriteString(conn, resp.SimpleString("OK"))
			} else {
				resp.InvalidReplicaCommand(conn)
			}
		case IsCommand(command, "PSYNC"):
			if !props.IsMaster() {
				resp.InvalidReplicaCommand(conn)
				continue
			}
			value := fmt.Sprintf("FULLRESYNC %s %d", props.ReplId(), props.ReplOffset())
			WriteString(conn, resp.SimpleString(value))
			WriteString(conn, resp.EmptyRdb())
			instance.LinkReplica(conn)
		default:
			WriteString(conn, resp.SimpleError("unknown command"))
		}
	}
}
