package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

var (
	wrongPasswordErr = errors.New("WRONGPASS invalid username-password pair or user is disabled.")
)

func authHandler(req *request.Request, args []string) error {
	if len(args) != 2 {
		return errors.New("Usage: AUTH <username> <password>")
	}

	username, password := args[0], args[1]

	user, ok := req.State.GetUser(username)
	if !ok {
		return wrongPasswordErr
	}

	if !user.ValidatePassword(password) {
		return wrongPasswordErr
	}

	req.State.WriteState(func(s *state.State) {
		s.User = user
	})

	return writeResponse(req, resp.NewString("OK"))
}