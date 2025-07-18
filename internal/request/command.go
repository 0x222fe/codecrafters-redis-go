package request

import (
	"errors"
	"fmt"
	"strings"

	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
)

type CommandKey string
type Command struct {
	Name CommandKey
	Args []string
}

func ParseCommandFromRESP(v resp.RESPValue) (Command, error) {
	arr, ok := v.GetArrayValue()
	if !ok {
		return Command{}, fmt.Errorf("expected RESP array, got %s", v.GetType())
	}

	if len(arr) < 1 {
		return Command{}, errors.New("command array must have at least one element")
	}

	cmdName, ok := arr[0].GetBulkStringValue()
	if !ok || cmdName == nil {
		return Command{}, errors.New("first element of command array must be a bulk string")
	}

	args := make([]string, 0, len(arr)-1)
	for _, v := range arr[1:] {
		arg, ok := v.GetBulkStringValue()
		if !ok {
			return Command{}, fmt.Errorf("command argument %v is not a string", v)
		}

		if arg == nil {
			return Command{}, errors.New("command argument cannot be nil")
		}

		args = append(args, *arg)
	}

	return Command{
		Name: CommandKey(strings.ToUpper(*cmdName)),
		Args: args,
	}, nil
}
