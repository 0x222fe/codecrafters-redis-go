package client

import (
	"bufio"
	"net"
	"sync"

	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/google/uuid"
)

type Client struct {
	ID     uuid.UUID
	mu     sync.Mutex
	conn   net.Conn
	writer *bufio.Writer
}

func NewClient(c net.Conn) *Client {
	return &Client{
		ID:     uuid.New(),
		conn:   c,
		writer: bufio.NewWriter(c),
	}
}

func (r *Client) WriteResp(resp resp.RESPValue) error {
	r.mu.Lock()
	r.mu.Unlock()

	_, err := r.Write(resp.Encode())
	return err
}

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
