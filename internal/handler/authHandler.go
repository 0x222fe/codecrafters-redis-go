package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

var (
	wrongPasswordErr = errors.New("WRONGPASS invalid username-password pair or user is disabled.")
)

func authHandler(req *request.Request, args []string) error {
	if len(args) != 2 {
		return errors.New("Usage: AUTH <username> <password>")
	}

	username, password := args[0], args[1]

	u, ok := req.State.GetUser(username)
	if !ok {
		return wrongPasswordErr
	}

	if !u.ValidatePassword(password) {
		return wrongPasswordErr
	}

	req.Client.SetUser(u)

	return writeResponse(req, resp.NewString("OK"))
}
