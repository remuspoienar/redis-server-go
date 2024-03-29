package resp

import (
	"fmt"
	"strconv"
	"strings"
)

func BulkString(input string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(input), input)
}

func SimpleString(str string) string {
	return fmt.Sprintf("+%s\r\n", str)
}

func SimpleError(str string) string {
	return fmt.Sprintf("-ERR %s\r\n", str)
}

func ParseCommand(input string) []string {
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
