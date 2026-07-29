package handler

import (
	"errors"
	"strings"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/user"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
)

var (
	noAuthErr = errors.New("NOAUTH Authentication required.")
)

func aclHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) < 1 {
		return errors.New("ACL requires at least 1 argument")
	}

	subcommand := strings.ToUpper(args[0])

	switch subcommand {
	case "WHOAMI":
		return aclWhoAmI(c, s, args[1:])
	case "GETUSER":
		return aclGetUser(c, s, args[1:])
	case "SETUSER":
		return aclSetUser(c, s, args[1:])
	default:
		return errors.New("unknown subcommand: " + subcommand)
	}
}

func aclWhoAmI(c *client.Client, s *state.AppState, args []string) error {
	if len(args) != 0 {
		return errors.New("ACL WHOAMI  requires no arguments")
	}

	u := c.Conn.User()
	if u == nil {
		return noAuthErr
	}

	name := u.Name()
	res := resp.NewBulkString(&name)
	return writeResponse(c, res)
}

func aclGetUser(c *client.Client, s *state.AppState, args []string) error {
	if len(args) != 1 {
		return errors.New("ACL GETUSER requires exactly 1 argument")
	}

	name := args[0]
	user, exists := s.GetUser(name)
	if !exists {
		return errors.New("ACL GETUSER: no such user")
	}

	arr := make([]resp.RESPValue, 0)

	nf := "flags"
	arr = append(arr, resp.NewBulkString(&nf))
	flags := user.Flags()
	arr = append(arr, resputil.BulkStringsToRESPArray(flags))

	np := "passwords"
	arr = append(arr, resp.NewBulkString(&np))
	passwords := user.Passwords()
	arr = append(arr, resputil.BulkStringsToRESPArray(passwords))

	result := resp.NewArray(arr)
	return writeResponse(c, result)
}

func aclSetUser(c *client.Client, s *state.AppState, args []string) error {
	if len(args) < 1 {
		return errors.New("ACL SETUSER requires at least 1 argument")
	}
	name, rules := args[0], args[1:]

	u, ok := s.GetUser(name)
	if !ok {
		u = user.New(name)
		s.AddUser(u)
	}

	for _, rule := range rules {
		switch {
		case strings.HasPrefix(rule, ">"):
			password := rule[1:]
			u.AddPassword(password)
		default:
			return errors.New("ACL SETUSER: unknown rule: " + rule)
		}

	}
	return writeResponse(c, resp.NewString("OK"))
}
