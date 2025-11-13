package handler

import (
	"errors"
	"strings"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func aclHandler(req *request.Request, args []string) error {
	if len(args) < 1 {
		return errors.New("ACL requires at least 1 argument")
	}

	subcommand := strings.ToUpper(args[0])

	switch subcommand {
	case "WHOAMI":
		return aclWhoami(req, args[1:])
	default:
		return errors.New("unknown subcommand: " + subcommand)
	}
}

func aclWhoami(req *request.Request, args []string) error {
	if len(args) != 0 {
		return errors.New("ACL WHOAMI  requires no arguments")
	}

	var user string
	req.State.ReadState(func(s state.State) {
		user = s.User
	})

	command := resp.NewBulkString(&user)
	return writeResponse(req, command)
}