package main

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/instances"
	. "github.com/codecrafters-io/redis-starter-go/app/internal"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"io"
	"net"
	"os"
	"strings"
	"time"
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

func handleConnection(conn net.Conn, i *instances.Instance) {
	defer CloseConnections(conn)

	props := (*i).Props()
	db := (*i).Db()

	for {

		if !i.Ready() && !i.IsPeer(conn) {
			fmt.Printf("[%s]Connection paused to finish a replica handshake\n", props.Role())
			time.Sleep(500 * time.Millisecond)
			continue
		}

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

		fmt.Printf("[%s] parsed command: `%s`\n", props.Role(), command)

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
				instance.PropagateCommand(buf[:n])
			}
			WriteString(conn, resp.SimpleString("OK"))
		case IsCommand(command, "INFO"):
			value := infoCommand(commandParts, props)
			WriteString(conn, resp.BulkString(value))
		case IsCommand(command, "REPLCONF"):
			if props.IsMaster() {
				if len(commandParts) >= 2 {
					if IsCommand(commandParts[1], "capa") || IsCommand(commandParts[1], "listening-port") {
						WriteString(conn, resp.SimpleString("OK"))
					}
					if IsCommand(commandParts[1], "ACK") {
						fmt.Println("Replica ACK complete")
					}
				}
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
			go func() { i.ReadyCh() <- false }()
			go func() {
				time.Sleep(500 * time.Millisecond)
				i.ReadyCh() <- true
			}()
		case IsCommand(command, "REPLCONF GETACK"):
			WriteString(conn, resp.Array("REPLCONF", "ACK", "0"))
			//instance.PropagateCommand([]byte(resp.Array("REPLCONF", "GETACK", "*")))
		default:
			WriteString(conn, resp.SimpleError("unknown command"))
		}
	}
}

var instance *instances.Instance

func main() {
	instance = instances.New()
	props := instance.Props()

	address := fmt.Sprintf("0.0.0.0:%d", props.Port())
	l, err := net.Listen("tcp", address)

	if err != nil {
		fmt.Println("Failed to bind port on", address)
		os.Exit(1)
	}
	defer CloseConnections(l)
	defer close(instance.ReadyCh())

	if props.IsMaster() {
		fmt.Printf("[%s]Server is listening on %s\n", props.Role(), address)

		go func() {
			time.Sleep(500 * time.Millisecond)
			instance.ReadyCh() <- true
		}()

	} else {
		go instance.ConnectToMaster()
		fmt.Printf("[%s]Server is listening on %s\nas a replica for master %s\n", props.Role(), address, props.MasterAddress())
	}

	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}
		go handleConnection(conn, instance)

	}
}
