package handler

import (
	"errors"
	"fmt"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils"
)

type commandType int
type commandSpec struct {
	handler commandHandler
	cmdType commandType
}
type commandHandler func(req *request.Request, args []string) error

func (h commandHandler) Handle(req *request.Request, cmd request.Command) error {
	return h(req, cmd.Args)
}

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
	MULTI    request.CommandKey = "MULTI"
	EXEC     request.CommandKey = "EXEC"
	DISCARD  request.CommandKey = "DISCARD"
	RPUSH    request.CommandKey = "RPUSH"
	LRANGE   request.CommandKey = "LRANGE"
)

var (
	handlerReg = map[request.CommandKey]commandSpec{
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
		MULTI:    {multiHandler, cmdTypeRead},
		RPUSH:    {rpushHandler, cmdTypeWrite},
		LRANGE:   {lrangeHandler, cmdTypeRead},
	}
)

func RunCommand(req *request.Request, cmd request.Command) error {
	cmdName := string(cmd.Name)

	if cmd.Name == EXEC {
		if !req.IsInTxn() {
			return errors.New("EXEC without MULTI")
		}

		resArr, executed, err := req.ExecTransaction()
		if err != nil {
			return fmt.Errorf("failed to execute transaction: %w", err)
		}

		if !executed {
			writeResponse(req, resp.RESPEmptyArray)
		} else {
			res := resp.NewRESPArray(resArr)
			writeResponse(req, res)
		}
		return nil
	}

	if cmd.Name == DISCARD {
		if !req.IsInTxn() {
			return errors.New("DISCARD without MULTI")
		}

		req.DiscardTransaction()
		writeResponse(req, resp.NewRESPString("OK"))
		return nil
	}

	spec, find := handlerReg[cmd.Name]
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

	if req.IsInTxn() {
		if cmd.Name == MULTI {
			return errors.New("MULTI calls can not be nested")
		}

		txnCmds := req.Transaction.Commands
		txnCmds = append(txnCmds, request.TxnCommand{Command: cmd, Handler: spec.handler})
		req.Transaction.Commands = txnCmds
		res := resp.NewRESPString("QUEUED")
		writeResponse(req, res)
		return nil
	}

	err := spec.handler(req, cmd.Args)
	if err != nil {
		return err
	}

	if spec.cmdType == cmdTypeWrite && !isReplica {
		replicaCommand := utils.StringsToRESPBulkStr(append([]string{cmdName}, cmd.Args...))
		encoded := replicaCommand.Encode()

		req.State.WriteState(func(s *state.State) {
			s.ReplicationOffset += len(encoded)
		})

		replicas := req.State.GetReplicas()

		for _, rep := range replicas {
			if _, err := rep.Client.Write(encoded); err != nil {
				fmt.Printf("failed to propagate command to replica %s: %v\n", rep.Client.RemoteAddr(), err)
			}
		}
	}

	return nil
}

func writeResponse(r *request.Request, res resp.RESPValue) error {
	writer := r.GetWriter()
	err := writer.WriteResp(res)
	if err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	fmt.Printf("Response sent to %s\n", r.Client.RemoteAddr())
	return nil
}
