package cnn

import (
	"bufio"
	"net"
	"sync"
)

type Connection struct {
	mu     sync.Mutex
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
}

func (r *Connection) Peek(n int) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.reader.Peek(n)
}

func (r *Connection) Read(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.reader.Read(p)
}

func (r *Connection) ReadByte() (byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.reader.ReadByte()
}

func (r *Connection) ReadString(delim byte) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.reader.ReadString(delim)
}

func (r *Connection) Discard(n int) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.reader.Discard(n)
}

func (r *Connection) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, err := r.writer.Write(p)
	if err != nil {
		return n, err
	}
	return n, r.writer.Flush()
}

func (r *Connection) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.conn.Close()
}

func (r *Connection) RemoteAddr() net.Addr {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.conn.RemoteAddr()
}
