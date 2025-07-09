package rdb

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/pkg/crc64"
)

func ReadRDBFile(filename string) (*RDB, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	rdb := &RDB{
		metadata:  make(map[string]string),
		databases: make(map[int]*database),
	}

	if err := parseHeader(reader, rdb); err != nil {
		return nil, fmt.Errorf("header parsing error: %v", err)
	}

	if err := parseMeta(reader, rdb); err != nil {
		return nil, fmt.Errorf("metadata parsing error: %v", err)
	}

	if err := parseDatabase(reader, rdb); err != nil {
		return nil, fmt.Errorf("database parsing error: %v", err)
	}

	checksum, err := parseEnd(reader)
	if err != nil {
		return nil, fmt.Errorf("end parsing error: %v", err)
	}

	file.Seek(0, io.SeekStart)

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	content := data[:len(data)-8]

	crc := crc64.Digest(content)
	actual := uint64(checksum[0]) |
		uint64(checksum[1])<<8 |
		uint64(checksum[2])<<16 |
		uint64(checksum[3])<<24 |
		uint64(checksum[4])<<32 |
		uint64(checksum[5])<<40 |
		uint64(checksum[6])<<48 |
		uint64(checksum[7])<<56

	if crc != actual {
		return nil, fmt.Errorf("checksum mismatch: computed %x, expected %x", crc, actual)
	}

	return rdb, nil
}

func parseHeader(reader *bufio.Reader, rdb *RDB) error {
	header := make([]byte, 9)
	if _, err := io.ReadFull(reader, header); err != nil {
		return err
	}

	if string(header[0:5]) != "REDIS" {
		return errors.New("invalid RDB file: missing REDIS signature")
	}

	rdb.version = string(header[5:9])
	return nil
}

func parseMeta(reader *bufio.Reader, rdb *RDB) error {
	for {
		flag, err := reader.Peek(1)
		if err != nil {
			return err
		}
		if flag[0] != metaFlag {
			return nil
		}

		_, err = reader.Discard(1)
		if err != nil {
			return err
		}

		name, err := readEncodedString(reader)
		if err != nil {
			return err
		}
		val, err := readEncodedString(reader)
		if err != nil {
			return err
		}
		rdb.metadata[name] = val
	}
}

func parseDatabase(reader *bufio.Reader, rdb *RDB) error {
	for {
		flag, err := reader.Peek(1)
		if err != nil {
			return err
		}
		if flag[0] == endFlag {
			return nil
		}

		if flag[0] != dbFlag {
			return fmt.Errorf("expected database flag, got 0x%02X", flag[0])
		}

		_, err = reader.Discard(1)
		if err != nil {
			return err
		}

		idx, err := readEncodedSize(reader)
		if err != nil {
			return err
		}

		_, has := rdb.databases[idx]
		if has {
			return fmt.Errorf("database with index %d already exists", idx)
		}

		flag, err = reader.Peek(1)
		if err != nil {
			return err
		}

		if flag[0] == dbFlag || flag[0] == endFlag {
			return nil
		}

		if flag[0] != tableFlag {
			return fmt.Errorf("expected table flag, got 0x%02X", flag[0])
		}
		_, err = reader.Discard(1)
		if err != nil {
			return err
		}

		hashSize, err := readEncodedSize(reader)
		if err != nil {
			return err
		}
		expirySize, err := readEncodedSize(reader)
		if err != nil {
			return err
		}

		db := &database{
			index:               idx,
			hashTableSize:       hashSize,
			expiryHashTableSize: expirySize,
			items:               make(map[string]*keyValue),
		}
		rdb.databases[idx] = db

		err = parseKeyValue(reader, db, hashSize, expirySize)
		if err != nil {
			return fmt.Errorf("error parsing key-value pairs: %v", err)
		}
	}
}

func parseKeyValue(reader *bufio.Reader, db *database, hashSize, expirySize int) error {

	expCount := 0
	for range hashSize {
		flag, err := reader.Peek(1)
		if err != nil {
			return err
		}

		hashSize--

		expiryAt := new(int64)

		if flag[0] == msExpFlag || flag[0] == sExpFlag {
			expCount++

			_, err = reader.Discard(1)
			if err != nil {
				return err
			}

			size := 4
			if flag[0] == msExpFlag {
				size = 8
			}

			bytes := make([]byte, size)
			if _, err := io.ReadFull(reader, bytes); err != nil {
				return err
			}

			if size == 4 {
				*expiryAt = int64(binary.LittleEndian.Uint32(bytes))
			} else {
				*expiryAt = int64(binary.LittleEndian.Uint64(bytes))
			}
		}

		typeByte, err := reader.ReadByte()
		if err != nil {
			return err
		}

		key, err := readEncodedString(reader)
		if err != nil {
			return err
		}
		val, err := readEncodedString(reader)
		if err != nil {
			return err
		}
		kv := &keyValue{
			key:       key,
			value:     val,
			valueType: typeByte,
			expireAt:  expiryAt,
		}

		db.items[key] = kv
	}

	if expCount != expirySize {
		return fmt.Errorf("expected %d expiry entries, got %d", expirySize, expCount)
	}

	return nil
}

func parseEnd(reader *bufio.Reader) (checksum []byte, err error) {
	flag, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	if flag != endFlag {
		return nil, fmt.Errorf("expected end flag, got 0x%02X", flag)
	}

	bytes := make([]byte, 8)
	if _, err := io.ReadFull(reader, bytes); err != nil {
		return nil, fmt.Errorf("error reading checksum: %v", err)
	}

	_, err = reader.ReadByte()
	if err == io.EOF {
		return bytes, nil
	}

	if err != nil {
		return nil, fmt.Errorf("unexpected error after checksum: %v", err)
	}
	return nil, fmt.Errorf("expected EOF after checksum, but found extra data")
}

func readEncodedSize(reader *bufio.Reader) (int, error) {
	firstByte, err := reader.ReadByte()
	if err != nil {
		return 0, err
	}

	firstTwoBits := firstByte >> 6

	mask := byte(0b_0011_1111)

	switch firstTwoBits {
	case 0b_00:
		return int(firstByte), nil
	case 0b_01:
		secondByte, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}

		return int(int16(firstByte&mask)<<8 | int16(secondByte)), nil
	case 0b_10:
		fourBytes := make([]byte, 4)
		if _, err := io.ReadFull(reader, fourBytes); err != nil {
			return 0, err
		}
		return int(binary.BigEndian.Uint32(fourBytes)), nil
	case 0b_11:
		fallthrough
	default:
		return 0, fmt.Errorf("invalid size encoding: first byte 0x%02X", firstByte)
	}
}

func readEncodedString(reader *bufio.Reader) (string, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return "", err
	}

	isNumString := b>>6 == 0b_11
	size := int(b)

	if isNumString {
		switch b {
		case 0xC0:
			size = 1
		case 0xC1:
			size = 2
		case 0xC2:
			size = 4
		case 0xC3:
			return "", errors.New("LZF compressed strings not supported")
		}
	}

	if size < 0 {
		return "", fmt.Errorf("invalid string size: %d", size)
	}

	bytes := make([]byte, size)
	_, err = io.ReadFull(reader, bytes)
	if err != nil {
		return "", err
	}

	if isNumString {
		switch b {
		case 0xC0: //INFO: int8
			return string(bytes), nil
		case 0xC1: // INFO: int16 little-endian
			return strconv.Itoa(int(binary.LittleEndian.Uint16(bytes))), nil
		case 0xC2: // INFO: int32 little-endian
			return strconv.Itoa(int(binary.LittleEndian.Uint32(bytes))), nil
		case 0xC3:
			return "", errors.New("LZF compressed strings not supported")
		}
	}
	return string(bytes), nil
}
