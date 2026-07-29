package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

var (
	wrongPasswordErr = errors.New("WRONGPASS invalid username-password pair or user is disabled.")
)

func authHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) != 2 {
		return errors.New("Usage: AUTH <username> <password>")
	}

	username, password := args[0], args[1]

	u, ok := s.GetUser(username)
	if !ok {
		return wrongPasswordErr
	}

	if !u.ValidatePassword(password) {
		return wrongPasswordErr
	}

	c.Conn.SetUser(u)

	return writeResponse(c, resp.NewString("OK"))
}
