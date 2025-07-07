package redis

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func Parse(reader *bufio.Reader) (string, []string, error) {
	l, err := reader.ReadString('\n')
	if err != nil {
		return "", nil, err
	}

	if l[0] != '*' {
		return "", nil, fmt.Errorf("invalid command format: expected array prefix '*', got %q", l[0])
	}

	numElements, err := strconv.Atoi(strings.TrimSpace(l[1:]))
	if err != nil {
		return "", nil, fmt.Errorf("invalid number of elements in command array: %w", err)
	}
	if numElements <= 0 {
		return "", nil, fmt.Errorf("invalid command: array cannot be empty or negative")
	}

	args := make([]string, numElements)
	for i := range numElements {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", nil, err
		}

		if line[0] != '$' {
			return "", nil, fmt.Errorf("invalid command format: expected bulk string prefix '$', got %q", line[0])
		}

		length, err := strconv.Atoi(strings.TrimSpace(line[1:]))
		if err != nil {
			return "", nil, fmt.Errorf("invalid bulk string length: %w", err)
		}

		data := make([]byte, length)
		_, err = io.ReadFull(reader, data)
		if err != nil {
			return "", nil, err
		}

		crlf := make([]byte, 2)
		_, err = io.ReadFull(reader, crlf)
		if err != nil {
			return "", nil, err
		}
		if string(crlf) != "\r\n" {
			return "", nil, fmt.Errorf("invalid command format: expected CRLF after bulk string, got %q", crlf)
		}

		args[i] = string(data)
	}

	commandName := strings.ToUpper(args[0])
	commandArgs := args[1:]

	return commandName, commandArgs, nil
}
