package instances

import (
	"fmt"
	. "github.com/codecrafters-io/redis-starter-go/app/internal"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"io"
	"net"
	"strings"
)

func handleConnection(conn net.Conn, i *Instance) {
	defer CloseConnections(conn)

	db := (*i).Db()
	for {
		buf := make([]byte, 4096)

		n, err := conn.Read(buf)

		if err != nil {
			if err == io.EOF {
				fmt.Println("[replication-handler] Connection closed by remote host")
				return
			}
			fmt.Println("[replication-handler] Error reading data", err.Error())
			continue
		}

		data := string(buf[:n])

		commandParts := resp.ParseCommand(data)
		command := strings.Join(commandParts, " ")

		fmt.Printf("parsed replication command: `%s`\n", command)
		fmt.Println("is ack cmd", IsCommand(command, "REPLCONF GETACK"))

		switch {
		case IsCommand(command, "SET"):
			px := ParsePX(command)
			db.Set(commandParts[1], commandParts[2], px)
		case IsCommand(command, "REPLCONF GETACK"):
			WriteString(conn, resp.Array("REPLCONF", "ACK", "0"))
		default:
			WriteString(conn, resp.SimpleString("OK"))
			//WriteString(conn, resp.SimpleError("unknown replication command"))
		}
	}

}
