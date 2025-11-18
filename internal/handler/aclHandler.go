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
	case "GETUSER":
		return aclGetUser(req, args[1:])
	default:
		return errors.New("unknown subcommand: " + subcommand)
	}
}

func aclWhoami(req *request.Request, args []string) error {
	if len(args) != 0 {
		return errors.New("ACL WHOAMI  requires no arguments")
	}

	var user *state.User
	req.State.ReadState(func(s state.State) {
		user = s.User
	})

	command := resp.NewBulkString(&user.Name)
	return writeResponse(req, command)
}

func aclGetUser(req *request.Request, args []string) error {
	if len(args) != 1 {
		return errors.New("ACL GETUSER requires exactly 1 argument")
	}

	name := args[0]
	user, exists := req.State.GetUser(name)
	if !exists {
		return errors.New("ACL GETUSER: no such user")
	}

	arr := make([]resp.RESPValue, 0)
	nf := "flags"
	arr = append(arr, resp.NewBulkString(&nf))

	flags := make([]resp.RESPValue, 0, len(user.Flags))
	for flag := range user.Flags {
		n := string(flag)
		flags = append(flags, resp.NewBulkString(&n))
	}
	arr = append(arr, resp.NewArray(flags))

	np := "passwords"
	arr = append(arr, resp.NewBulkString(&np))

	passwords := make([]resp.RESPValue, 0, len(user.Passwords))
	for password := range user.Passwords {
		passwords = append(passwords, resp.NewBulkString(&password))
	}
	arr = append(arr, resp.NewArray(passwords))

	result := resp.NewArray(arr)
	return writeResponse(req, result)
}
