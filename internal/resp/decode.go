package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func DecodeRESPInput(r *bufio.Reader) (val RESPValue, bytes int, err error) {
	reader := newCountingReader(r)
	val, err = decodeRESPInput(reader)
	if err != nil {
		return RESPValue{}, 0, err
	}
	return val, reader.count, nil
}

func DecodeRESPInputExact(r *bufio.Reader, valType respValueType) (val RESPValue, bytes int, err error) {
	reader := newCountingReader(r)

	val, err = decodeRESPInputExact(reader, valType)

	if err != nil {
		return RESPValue{}, 0, err
	}
	return val, reader.count, nil
}

func decodeRESPInput(reader *countingReader) (RESPValue, error) {
	flag, err := reader.Peek(1)
	if err != nil {
		return RESPValue{}, err
	}

	switch flag[0] {
	case '+':
		return parseStr(reader)
	case '-':
		return parseErr(reader)
	case '$':
		return parseBulkStr(reader)
	case ':':
		return parseInt(reader)
	case '*':
		return parseArray(reader)
	default:
		return RESPValue{}, fmt.Errorf("unknown RESP type: %q", flag[0])
	}
}

func decodeRESPInputExact(reader *countingReader, valType respValueType) (RESPValue, error) {
	switch valType {
	case RESPStr:
		return parseStr(reader)
	case RESPErr:
		return parseErr(reader)
	case RESPBulkStr:
		return parseBulkStr(reader)
	case RESPInt:
		return parseInt(reader)
	case RESPArr:
		return parseArray(reader)
	default:
		return RESPValue{}, fmt.Errorf("unknown RESP value type: %d", valType)
	}
}

func parseStr(reader *countingReader) (RESPValue, error) {
	flag, err := reader.ReadByte()
	if err != nil {
		return RESPValue{}, err
	}
	if flag != '+' {
		return RESPValue{}, fmt.Errorf("expected '+' for RESP string, got %q", flag)
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		return RESPValue{}, err
	}
	s := strings.TrimSuffix(line, "\r\n")
	return NewRESPString(s), nil
}

func parseErr(reader *countingReader) (RESPValue, error) {
	flag, err := reader.ReadByte()
	if err != nil {
		return RESPValue{}, err
	}
	if flag != '-' {
		return RESPValue{}, fmt.Errorf("expected '-' for RESP error, got %q", flag)
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		return RESPValue{}, err
	}
	s := strings.TrimSuffix(line, "\r\n")
	return NewRESPString(s), nil
}

func parseBulkStr(reader *countingReader) (RESPValue, error) {
	flag, err := reader.ReadByte()
	if err != nil {
		return RESPValue{}, err
	}
	if flag != '$' {
		return RESPValue{}, fmt.Errorf("expected '$' for RESP bulk string, got %q", flag)
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		return RESPValue{}, err
	}
	length, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		return RESPValue{}, err
	}
	if length == -1 {
		return NewRESPBulkString(nil), nil
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(reader, data); err != nil {
		return RESPValue{}, err
	}

	//INFO: discard CRLF after bulk string data required by RESP protocol
	if _, err := reader.Discard(2); err != nil {
		return RESPValue{}, err
	}
	s := string(data)
	return NewRESPBulkString(&s), nil
}

func parseInt(reader *countingReader) (RESPValue, error) {
	flag, err := reader.ReadByte()
	if err != nil {
		return RESPValue{}, err
	}
	if flag != ':' {
		return RESPValue{}, fmt.Errorf("expected ':' for RESP integer, got %q", flag)
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return RESPValue{}, err
	}
	i, err := strconv.ParseInt(strings.TrimSpace(line), 10, 64)
	if err != nil {
		return RESPValue{}, err
	}
	return NewRESPInt(i), nil
}

func parseArray(reader *countingReader) (RESPValue, error) {
	flag, err := reader.ReadByte()

	if err != nil {
		return RESPValue{}, err
	}
	if flag != '*' {
		return RESPValue{}, fmt.Errorf("expected '*' for RESP array, got %q", flag)
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return RESPValue{}, err
	}
	length, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		return RESPValue{}, err
	}

	if length == -1 {
		return NewRESPArray(nil), nil
	}

	arr := make([]RESPValue, length)
	for i := range length {
		v, err := decodeRESPInput(reader)
		if err != nil {
			return RESPValue{}, err
		}
		arr[i] = v
	}
	return NewRESPArray(arr), nil
}