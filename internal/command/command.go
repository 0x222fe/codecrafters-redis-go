package command

import (
	"errors"
	"fmt"
	"io"

	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
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
		replicaCommand, err := resp.RESPEncode(append([]string{cmdName}, cmd.Args...))
		if err != nil {
			return fmt.Errorf("failed to encode command for replicas: %w", err)
		}
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
