package command

import (
	"errors"
	"fmt"
	"io"

	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

type command string
type commandType int
type commandSpec struct {
	handler commandHandler
	cmdType commandType
}
type commandHandler func(state *state.AppState, args []string, writer io.Writer) error

type contextKey string

var (
	ConnectionContextKey contextKey = "conn"
)

const (
	cmdTypeRead commandType = iota
	cmdTypeWrite
)

const (
	PING     command = "PING"
	ECHO     command = "ECHO"
	SET      command = "SET"
	GET      command = "GET"
	CONFIG   command = "CONFIG"
	KEYS     command = "KEYS"
	INFO     command = "INFO"
	REPLCONF command = "REPLCONF"
	PSYNC    command = "PSYNC"
)

var (
	commands = map[command]commandSpec{
		PING:     {pingHandler, cmdTypeRead},
		ECHO:     {echoHandler, cmdTypeRead},
		SET:      {setHandler, cmdTypeWrite},
		GET:      {getHandler, cmdTypeRead},
		CONFIG:   {configHandler, cmdTypeRead},
		KEYS:     {keysHandler, cmdTypeRead},
		INFO:     {infoHandler, cmdTypeRead},
		REPLCONF: {replconfHandler, cmdTypeRead},
		PSYNC:    {psyncHandler, cmdTypeRead},
	}
)

func RunCommand(appState *state.AppState, cmd string, args []string, writer io.Writer) error {
	spec, exists := commands[command(cmd)]
	if !exists {
		return errors.New("unknown command: " + cmd)
	}

	if spec.cmdType == cmdTypeWrite {
		var isReplica bool
		appState.ReadState(func(s state.State) {
			isReplica = s.IsReplica
		})

		if isReplica {
			return errors.New("replica cannot execute write commands")
		}
	}

	err := spec.handler(appState, args, writer)
	if err != nil {
		return err
	}

	if spec.cmdType == cmdTypeWrite {
		replicaCommand, err := resp.RESPEncode(append([]string{cmd}, args...))
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
