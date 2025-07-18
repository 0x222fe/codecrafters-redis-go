package command

import (
	"errors"
	"fmt"
	"strings"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils"
)

type CommandKey string
type commandType int
type Command struct {
	Name CommandKey
	Args []string
}
type commandSpec struct {
	handler commandHandler
	cmdType commandType
}
type commandHandler func(req *request.Request, args []string) error

type contextKey string

var (
	ConnectionContextKey contextKey = "conn"
)

const (
	cmdTypeRead commandType = iota
	cmdTypeWrite
)

const (
	PING     CommandKey = "PING"
	ECHO     CommandKey = "ECHO"
	SET      CommandKey = "SET"
	GET      CommandKey = "GET"
	CONFIG   CommandKey = "CONFIG"
	KEYS     CommandKey = "KEYS"
	INFO     CommandKey = "INFO"
	REPLCONF CommandKey = "REPLCONF"
	PSYNC    CommandKey = "PSYNC"
	WAIT     CommandKey = "WAIT"
	TYPE     CommandKey = "TYPE"
	XADD     CommandKey = "XADD"
	XRANGE   CommandKey = "XRANGE"
	XREAD    CommandKey = "XREAD"
	INCR     CommandKey = "INCR"
)

var (
	commands = map[CommandKey]commandSpec{
		PING:     {pingHandler, cmdTypeRead},
		ECHO:     {echoHandler, cmdTypeRead},
		SET:      {setHandler, cmdTypeWrite},
		GET:      {getHandler, cmdTypeRead},
		CONFIG:   {configHandler, cmdTypeRead},
		KEYS:     {keysHandler, cmdTypeRead},
		INFO:     {infoHandler, cmdTypeRead},
		REPLCONF: {replconfHandler, cmdTypeRead},
		PSYNC:    {psyncHandler, cmdTypeRead},
		WAIT:     {waitHandler, cmdTypeRead},
		TYPE:     {typeHandler, cmdTypeRead},
		XADD:     {xaddHandler, cmdTypeWrite},
		XRANGE:   {xrangeHandler, cmdTypeRead},
		XREAD:    {xreadHandler, cmdTypeRead},
		INCR:     {incrHandler, cmdTypeWrite},
	}
)

func FromRESP(v resp.RESPValue) (Command, error) {
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

func RunCommand(req *request.Request, cmd Command) error {
	cmdName := string(cmd.Name)

	spec, find := commands[cmd.Name]
	if !find {
		return errors.New("unknown command: " + cmdName)
	}

	var isReplica bool
	req.State.ReadState(func(s state.State) {
		isReplica = s.IsReplica
	})

	if spec.cmdType == cmdTypeWrite &&
		isReplica &&
		!req.Propagated {
		return errors.New("replica cannot execute write commands")
	}

	err := spec.handler(req, cmd.Args)
	if err != nil {
		return err
	}

	if spec.cmdType == cmdTypeWrite && !isReplica {
		replicaCommand := utils.EncodeBulkStrArrToRESP(append([]string{cmdName}, cmd.Args...))

		req.State.WriteState(func(s *state.State) {
			s.ReplicationOffset += len(replicaCommand)
		})

		replicas := req.State.GetReplicas()

		for _, rep := range replicas {
			if _, err := rep.Client.Write(replicaCommand); err != nil {
				fmt.Printf("failed to propagate command to replica %s: %v\n", rep.Client.RemoteAddr(), err)
			}
		}
	}

	return nil
}

func writeResponse(c *client.Client, response []byte) error {
	_, err := c.Write(response)
	fmt.Printf("Response sent to %s: %q\n", c.RemoteAddr(), string(response))
	if err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	return nil
}
