package resp

import (
	"bufio"
)

type countingReader struct {
	reader *bufio.Reader
	count  int
}

func newCountingReader(reader *bufio.Reader) *countingReader {
	return &countingReader{
		reader: reader,
		count:  0,
	}
}

func (r *countingReader) Peek(n int) ([]byte, error) {
	data, err := r.reader.Peek(n)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 {
		r.count += n
	}
	return n, err
}

func (r *countingReader) ReadByte() (byte, error) {
	b, err := r.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	r.count++
	return b, nil
}

func (r *countingReader) ReadString(delim byte) (string, error) {
	str, err := r.reader.ReadString(delim)
	if err == nil {
		r.count += len(str)
	}
	return str, err
}

func (r *countingReader) Discard(n int) (int, error) {
	n, err := r.reader.Discard(n)
	if n > 0 {
		r.count += n
	}
	return n, err
}
