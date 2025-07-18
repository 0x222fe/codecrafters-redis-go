package handler

import (
	"errors"
	"fmt"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils"
)

type commandType int
type commandSpec struct {
	handler commandHandler
	cmdType commandType
}
type commandHandler func(req *request.Request, args []string) error

const (
	cmdTypeRead commandType = iota
	cmdTypeWrite
)

const (
	PING     request.CommandKey = "PING"
	ECHO     request.CommandKey = "ECHO"
	SET      request.CommandKey = "SET"
	GET      request.CommandKey = "GET"
	CONFIG   request.CommandKey = "CONFIG"
	KEYS     request.CommandKey = "KEYS"
	INFO     request.CommandKey = "INFO"
	REPLCONF request.CommandKey = "REPLCONF"
	PSYNC    request.CommandKey = "PSYNC"
	WAIT     request.CommandKey = "WAIT"
	TYPE     request.CommandKey = "TYPE"
	XADD     request.CommandKey = "XADD"
	XRANGE   request.CommandKey = "XRANGE"
	XREAD    request.CommandKey = "XREAD"
	INCR     request.CommandKey = "INCR"
)

var (
	commands = map[request.CommandKey]commandSpec{
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

func RunCommand(req *request.Request, cmd request.Command) error {
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
