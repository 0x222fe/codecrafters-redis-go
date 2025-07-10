package command

import (
	"errors"
	"fmt"
	"io"

	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

type command string
type commandType int
type commandSpec struct {
	handler commandHandler
	cmdType commandType
}
type commandHandler func(state *state.AppState, args []string, writer io.Writer) error

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

	var isReplica bool
	appState.ReadState(func(s state.State) {
		isReplica = s.IsReplica
	})

	if spec.cmdType == cmdTypeWrite && isReplica {
		return errors.New("replica cannot execute write commands")
	}

	return spec.handler(appState, args, writer)
}

func writeResponse(writer io.Writer, response []byte) error {
	_, err := writer.Write(response)
	if err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	return nil
}
