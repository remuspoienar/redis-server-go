package instances

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/internal"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"io"
	"net"
	"strings"
)

func handleConnection(conn net.Conn, i *Instance) {
	defer internal.CloseConnections(conn)

	db := (*i).Db()
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

		case internal.IsCommand(command, "SET"):
			px := internal.ParsePX(command)
			db.Set(commandParts[1], commandParts[2], px)

		default:
			internal.WriteString(conn, resp.SimpleError("unknown command"))
		}
	}

}
