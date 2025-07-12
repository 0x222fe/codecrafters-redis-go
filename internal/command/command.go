package command

import (
	"errors"
	"fmt"
	"io"

	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils"
)

type CommandKey string
type commandType int
type Command struct {
	Name CommandKey
	Args []string
	// Wether this command is propagated from master
	Propagated bool
}
type simpleCommandSpec struct {
	handler simpleHandler
	cmdType commandType
}
type streamCommandSpec struct {
	handler streamHandler
	cmdType commandType
}
type simpleHandler func(state *state.AppState, args []string) ([]byte, error)
type streamHandler func(state *state.AppState, args []string, writer io.Writer) error

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
)

var (
	simpleCommands = map[CommandKey]simpleCommandSpec{
		PING:     {pingHandler, cmdTypeRead},
		ECHO:     {echoHandler, cmdTypeRead},
		SET:      {setHandler, cmdTypeWrite},
		GET:      {getHandler, cmdTypeRead},
		CONFIG:   {configHandler, cmdTypeRead},
		KEYS:     {keysHandler, cmdTypeRead},
		INFO:     {infoHandler, cmdTypeRead},
		REPLCONF: {replconfHandler, cmdTypeRead},
	}
	streamCommands = map[CommandKey]streamCommandSpec{
		PSYNC: {psyncHandler, cmdTypeRead},
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
		Name:       CommandKey(*cmdName),
		Args:       args,
		Propagated: false,
	}, nil
}

func RunCommand(appState *state.AppState, cmd Command, writer io.Writer) error {
	cmdName := string(cmd.Name)

	simpleSpec, findInSimple := simpleCommands[cmd.Name]
	streamSpec, findInStream := streamCommands[cmd.Name]

	if !findInSimple && !findInStream {
		return errors.New("unknown command: " + cmdName)
	}

	var commandType commandType
	if findInSimple {
		commandType = simpleSpec.cmdType
	} else {
		commandType = streamSpec.cmdType
	}

	if commandType == cmdTypeWrite && !cmd.Propagated {
		var isReplica bool
		appState.ReadState(func(s state.State) {
			isReplica = s.IsReplica
		})

		if isReplica {
			return errors.New("replica cannot execute write commands")
		}
	}

	if findInSimple {
		bytes, err := simpleSpec.handler(appState, cmd.Args)
		if err != nil {
			return err
		}

		if !cmd.Propagated {
			err := writeResponse(writer, bytes)
			if err != nil {
				return err
			}
		}
	} else {
		if cmd.Propagated {
			return errors.New("stream commands cannot be propagated")
		}

		err := streamSpec.handler(appState, cmd.Args, writer)
		if err != nil {
			return err
		}
	}

	if commandType == cmdTypeWrite {
		replicaCommand := utils.EncodeStringSliceToRESP(append([]string{cmdName}, cmd.Args...))

		appState.IterateReplicas(func(conn io.Writer) {
			if _, err := conn.Write(replicaCommand); err != nil {
				fmt.Printf("failed to send command to replica: %v\n", err)
			}
		})
	}

	return nil
}

func writeResponse(writer io.Writer, response []byte) error {
	_, err := writer.Write(response)
	if err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	return nil
}
