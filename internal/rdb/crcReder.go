package rdb

import (
	"bufio"
	"hash"
	"io"

	"github.com/0x222fe/codecrafters-redis-go/pkg/crc64"
)

type crcReader struct {
	reader *bufio.Reader
	crc    hash.Hash64
}

func newCRCReader(reader io.Reader) *crcReader {
	return &crcReader{
		reader: bufio.NewReader(reader),
		crc:    crc64.New(),
	}
}

func (r *crcReader) Peek(n int) ([]byte, error) {
	data, err := r.reader.Peek(n)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *crcReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 {
		r.crc.Write(p[:n])
	}
	return n, err
}

func (r *crcReader) ReadByte() (byte, error) {
	b, err := r.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	r.crc.Write([]byte{b})
	return b, nil
}

func (r *crcReader) Discard(n int) (int, error) {
	copied, err := io.CopyN(r.crc, r.reader, int64(n))
	return int(copied), err
}
