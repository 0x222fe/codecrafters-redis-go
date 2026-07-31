package connection

import (
	"bufio"
	"net"
	"sync"

	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/user"
	"github.com/google/uuid"
)

type Connection struct {
	ID      uuid.UUID
	mu      sync.Mutex
	rawConn net.Conn
	writer  *bufio.Writer
	user    *user.User
}

func NewConnection(rawConn net.Conn, u *user.User) *Connection {
	return &Connection{
		ID:      uuid.New(),
		rawConn: rawConn,
		writer:  bufio.NewWriter(rawConn),
		user:    u,
	}
}

func (conn *Connection) User() *user.User {
	return conn.user
}

func (conn *Connection) SetUser(user *user.User) {
	conn.user = user
}

func (conn *Connection) WriteResp(resp resp.RESPValue) error {
	_, err := conn.Write(resp.Bytes())
	return err
}

func (conn *Connection) Write(p []byte) (int, error) {
	conn.mu.Lock()
	writer := conn.writer
	conn.mu.Unlock()

	n, err := writer.Write(p)
	if err != nil {
		return n, err
	}
	return n, writer.Flush()
}

func (conn *Connection) Close() error {
	conn.mu.Lock()
	rawConn := conn.rawConn
	conn.mu.Unlock()
	return rawConn.Close()
}

func (conn *Connection) RemoteAddr() net.Addr {
	conn.mu.Lock()
	rawConn := conn.rawConn
	conn.mu.Unlock()
	return rawConn.RemoteAddr()
}
