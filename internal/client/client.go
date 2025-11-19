package client

import (
	"bufio"
	"net"
	"sync"

	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/user"
	"github.com/google/uuid"
)

type Client struct {
	ID     uuid.UUID
	mu     sync.Mutex
	conn   net.Conn
	writer *bufio.Writer
	user   *user.User
}

func NewClient(c net.Conn) *Client {
	return &Client{
		ID:     uuid.New(),
		conn:   c,
		writer: bufio.NewWriter(c),
		user:   user.DefaulUser,
	}
}

func (r *Client) User() *user.User {
	return r.user
}

func (r *Client) SetUser(user *user.User) {
	r.user = user
}

func (r *Client) WriteResp(resp resp.RESPValue) error {
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
