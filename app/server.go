package main

import (
	"fmt"
	"io"
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
const DOCS = "COMMAND DOCS"

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

		//fmt.Println("Incoming data:\n\n" + data)

		command := strings.Join(fromResp3(data), " ")

		fmt.Printf("parsed command: `%s`\n", command)

		if isCommand(command, PING) {
			conn.Write([]byte(lineToResp3(PONG)))
		} else if isCommand(command, DOCS) {
			conn.Write([]byte(lineToResp3("OK")))
		} else {
			conn.Write([]byte(errorToResp3("unknown command")))
		}
	}
}

func isCommand(input string, value string) bool {
	return strings.ToUpper(input) == value
}

func lineToResp3(str string) string {
	return fmt.Sprintf("+%s\r\n", str)
}

func errorToResp3(str string) string {
	return fmt.Sprintf("-ERR %s\r\n", str)
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
