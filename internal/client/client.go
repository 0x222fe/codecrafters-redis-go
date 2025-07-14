package client

import (
	"bufio"
	"net"
	"sync"

	"github.com/google/uuid"
)

type Client struct {
	ID   uuid.UUID
	mu   sync.Mutex
	conn net.Conn
	// reader *bufio.Reader
	writer *bufio.Writer
}

func NewClient(c net.Conn) *Client {
	return &Client{
		ID:   uuid.New(),
		conn: c,
		// reader: bufio.NewReader(c),
		writer: bufio.NewWriter(c),
	}
}

// func (r *Client) Peek(n int) ([]byte, error) {
// 	r.mu.Lock()
// 	reader := r.reader
// 	r.mu.Unlock()
// 	return reader.Peek(n)
// }
//
// func (r *Client) Read(p []byte) (int, error) {
// 	r.mu.Lock()
// 	reader := r.reader
// 	r.mu.Unlock()
// 	return reader.Read(p)
// }
//
// func (r *Client) ReadByte() (byte, error) {
// 	r.mu.Lock()
// 	reader := r.reader
// 	r.mu.Unlock()
// 	return reader.ReadByte()
// }
//
// func (r *Client) ReadString(delim byte) (string, error) {
// 	r.mu.Lock()
// 	reader := r.reader
// 	r.mu.Unlock()
// 	return reader.ReadString(delim)
// }
//
// func (r *Client) Discard(n int) (int, error) {
// 	r.mu.Lock()
// 	reader := r.reader
// 	r.mu.Unlock()
// 	return reader.Discard(n)
// }

func (r *Client) Write(p []byte) (int, error) {
	r.mu.Lock()
	writer := r.writer
	r.mu.Unlock()

	n, err := writer.Write(p)
	if err != nil {
		return n, err
	}
	return n, writer.Flush()
}

func (r *Client) Close() error {
	r.mu.Lock()
	conn := r.conn
	r.mu.Unlock()
	return conn.Close()
}

func (r *Client) RemoteAddr() net.Addr {
	r.mu.Lock()
	conn := r.conn
	r.mu.Unlock()
	return conn.RemoteAddr()
}
