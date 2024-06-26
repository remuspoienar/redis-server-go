package resp

import (
	"encoding/hex"
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/internal"
	"net"
	"strconv"
	"strings"
)

func BulkString(input any) string {
	if input == nil {
		return "$-1\r\n"
	}
	str := fmt.Sprintf("%s", input)
	return fmt.Sprintf("$%d\r\n%s\r\n", len(str), str)
}

func SimpleString(input any) string {
	return fmt.Sprintf("+%s\r\n", input)
}

func SimpleError(str string) string {
	return fmt.Sprintf("-ERR %s\r\n", str)
}

func Array(input ...string) string {
	res := fmt.Sprintf("*%d\r\n", len(input))
	for _, str := range input {
		res += BulkString(str)
	}
	return res
}

func ParseCommand(input string) []string {
	if strings.HasPrefix(input, "$") {
		return []string{input}
	}

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

func EmptyRdb() string {
	bin, _ := hex.DecodeString(internal.EmptyRdbHex)
	res := BulkString(bin)
	return strings.TrimRight(res, "\r\n")
}

func InvalidReplicaCommand(conn net.Conn) {
	internal.WriteString(conn, SimpleError("This command cannot be executed on a replica"))
}
