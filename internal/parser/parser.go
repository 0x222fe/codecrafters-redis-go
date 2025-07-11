package parser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/0x222fe/codecrafters-redis-go/internal/command"
)

var (
	cmdZero = command.Command{}
)

func Parse(reader *bufio.Reader) (command.Command, error) {
	l, err := reader.ReadString('\n')
	if err != nil {
		return cmdZero, err
	}

	if l[0] != '*' {
		return cmdZero, fmt.Errorf("invalid command format: expected array prefix '*', got %q", l[0])
	}

	numElements, err := strconv.Atoi(strings.TrimSpace(l[1:]))
	if err != nil {
		return cmdZero, fmt.Errorf("invalid number of elements in command array: %w", err)
	}
	if numElements <= 0 {
		return cmdZero, fmt.Errorf("invalid command: array cannot be empty or negative")
	}

	args := make([]string, numElements)
	for i := range numElements {
		line, err := reader.ReadString('\n')
		if err != nil {
			return cmdZero, err
		}

		if line[0] != '$' {
			return cmdZero, fmt.Errorf("invalid command format: expected bulk string prefix '$', got %q", line[0])
		}

		length, err := strconv.Atoi(strings.TrimSpace(line[1:]))
		if err != nil {
			return cmdZero, fmt.Errorf("invalid bulk string length: %w", err)
		}

		data := make([]byte, length)
		_, err = io.ReadFull(reader, data)
		if err != nil {
			return cmdZero, err
		}

		crlf := make([]byte, 2)
		_, err = io.ReadFull(reader, crlf)
		if err != nil {
			return cmdZero, err
		}
		if string(crlf) != "\r\n" {
			return cmdZero, fmt.Errorf("invalid command format: expected CRLF after bulk string, got %q", crlf)
		}

		args[i] = string(data)
	}

	cmd := command.Command{
		Name: command.CommandKey(strings.ToUpper(args[0])),
		Args: args[1:],
	}

	return cmd, nil
}
