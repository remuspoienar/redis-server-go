package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	// Uncomment this block to pass the first stage
	// "net"
	// "os"
)

const PORT = "6379"
const PING = "PING"
const PONG = "PONG"

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

	buf := make([]byte, 1024)

	n, err := conn.Read(buf)

	if err != nil {
		fmt.Println("Error reading data", err.Error())
	}

	data := string(buf[:n])

	fmt.Println("Incoming data:\n\n" + data)

	command := fromResp3("*1\r\n$4\r\nping\r\n")

	if strings.ToUpper(command[0]) == PING {
		conn.Write([]byte(toResp3(PONG)))
		return
	}

	conn.Write([]byte(toResp3("OK")))
}

func toResp3(str string) string {
	return fmt.Sprintf("+%s\r\n", str)
}

func fromResp3(input string) []string {
	lines := strings.Split(strings.Trim(input, "\r\n"), "\r\n")
	argsNoStr := strings.Split(lines[0], "*")[1]
	argsNo, _ := strconv.ParseInt(argsNoStr, 10, 32)

	var args []string

	for i := 1; i <= 2*int(argsNo); i += 2 {
		argLenStr := strings.Split(lines[i], "$")[1]
		argLen, _ := strconv.ParseInt(argLenStr, 10, 32)
		arg := lines[i+1][:argLen]
		args = append(args, arg)
	}

	return args
}
