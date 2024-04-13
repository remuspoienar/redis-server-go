package internal

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const EmptyRdbHex = "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"

func IsCommand(input string, value string) bool {
	return strings.HasPrefix(strings.ToUpper(input), value)
}

func WriteString(conn net.Conn, msg string) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("Error writing data")
	}
}

func ParsePX(command string) int64 {
	var pxRaw string

	pxArgs := strings.Split(strings.ToUpper(command), "PX")
	if len(pxArgs) == 1 {
		pxRaw = "-1"
	} else {
		pxRaw = strings.TrimSpace(pxArgs[1])
	}
	px, err := strconv.ParseInt(pxRaw, 10, 32)
	if err != nil {
		fmt.Printf("Could not parse PX from command `%s`\n", command)
		px = -1
	}
	return px
}

func CloseConnections(closable any) {
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
