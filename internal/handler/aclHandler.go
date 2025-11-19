package handler

import (
	"errors"
	"strings"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
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
	case "SETUSER":
		return aclSetUser(req, args[1:])
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

	name := user.GetName()
	command := resp.NewBulkString(&name)
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
	flags := user.GetFlags()
	arr = append(arr, resputil.BulkStringsToRESPArray(flags))

	np := "passwords"
	arr = append(arr, resp.NewBulkString(&np))
	passwords := user.GetPasswords()
	arr = append(arr, resputil.BulkStringsToRESPArray(passwords))

	result := resp.NewArray(arr)
	return writeResponse(req, result)
}

func aclSetUser(req *request.Request, args []string) error {
	if len(args) < 1 {
		return errors.New("ACL SETUSER requires at least 1 argument")
	}
	name, rules := args[0], args[1:]

	user, ok := req.State.GetUser(name)
	if !ok {
		user = state.NewUser(name)
		req.State.AddUser(user)
	}

	for _, rule := range rules {
		switch {
		case strings.HasPrefix(rule, ">"):
			password := rule[1:]
			user.AddPassword(password)
		default:
			return errors.New("ACL SETUSER: unknown rule: " + rule)
		}

	}
	return writeResponse(req, resp.NewString("OK"))
}