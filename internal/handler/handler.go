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
	handler          commandHandler
	cmdType          commandType
	allowedInSubMode bool
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
	PING        request.CommandKey = "PING"
	ECHO        request.CommandKey = "ECHO"
	SET         request.CommandKey = "SET"
	GET         request.CommandKey = "GET"
	CONFIG      request.CommandKey = "CONFIG"
	KEYS        request.CommandKey = "KEYS"
	INFO        request.CommandKey = "INFO"
	REPLCONF    request.CommandKey = "REPLCONF"
	PSYNC       request.CommandKey = "PSYNC"
	WAIT        request.CommandKey = "WAIT"
	TYPE        request.CommandKey = "TYPE"
	XADD        request.CommandKey = "XADD"
	XRANGE      request.CommandKey = "XRANGE"
	XREAD       request.CommandKey = "XREAD"
	INCR        request.CommandKey = "INCR"
	MULTI       request.CommandKey = "MULTI"
	EXEC        request.CommandKey = "EXEC"
	DISCARD     request.CommandKey = "DISCARD"
	LPUSH       request.CommandKey = "LPUSH"
	RPUSH       request.CommandKey = "RPUSH"
	LRANGE      request.CommandKey = "LRANGE"
	LLEN        request.CommandKey = "LLEN"
	LPOP        request.CommandKey = "LPOP"
	BLPOP       request.CommandKey = "BLPOP"
	RPOP        request.CommandKey = "RPOP"
	SUBSCRIBE   request.CommandKey = "SUBSCRIBE"
	UNSUBSCRIBE request.CommandKey = "UNSUBSCRIBE"
	PUBLISH     request.CommandKey = "PUBLISH"
	ZADD        request.CommandKey = "ZADD"
	ZRANK       request.CommandKey = "ZRANK"
	ZRANGE      request.CommandKey = "ZRANGE"
	ZCARD       request.CommandKey = "ZCARD"
	ZSCORE      request.CommandKey = "ZSCORE"
)

var (
	handlerReg = map[request.CommandKey]commandSpec{
		PING:        {handler: pingHandler, allowedInSubMode: true},
		ECHO:        {handler: echoHandler},
		SET:         {handler: setHandler, cmdType: cmdTypeWrite},
		GET:         {handler: getHandler},
		CONFIG:      {handler: configHandler},
		KEYS:        {handler: keysHandler},
		INFO:        {handler: infoHandler},
		REPLCONF:    {handler: replconfHandler},
		PSYNC:       {handler: psyncHandler},
		WAIT:        {handler: waitHandler},
		TYPE:        {handler: typeHandler},
		XADD:        {handler: xaddHandler, cmdType: cmdTypeWrite},
		XRANGE:      {handler: xrangeHandler},
		XREAD:       {handler: xreadHandler},
		INCR:        {handler: incrHandler, cmdType: cmdTypeWrite},
		MULTI:       {handler: multiHandler},
		LPUSH:       {handler: lpushHandler, cmdType: cmdTypeWrite},
		RPUSH:       {handler: rpushHandler, cmdType: cmdTypeWrite},
		LRANGE:      {handler: lrangeHandler},
		LLEN:        {handler: llenHandler},
		LPOP:        {handler: lpopHandler, cmdType: cmdTypeWrite},
		BLPOP:       {handler: blpopHandler, cmdType: cmdTypeWrite},
		RPOP:        {handler: rpopHandler, cmdType: cmdTypeWrite},
		SUBSCRIBE:   {handler: subscribeHandler, allowedInSubMode: true},
		UNSUBSCRIBE: {handler: unsubscribeHandler, allowedInSubMode: true},
		PUBLISH:     {handler: publishHandler, cmdType: cmdTypeWrite, allowedInSubMode: true},
		ZADD:        {handler: zaddHandler, cmdType: cmdTypeWrite},
		ZRANK:       {handler: zrankHandler, cmdType: cmdTypeRead},
		ZRANGE:      {handler: zrangeHandler, cmdType: cmdTypeRead},
		ZCARD:       {handler: zcardHandler, cmdType: cmdTypeRead},
		ZSCORE:      {handler: zscoreHandler, cmdType: cmdTypeRead},
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
			res := resp.NewArray(resArr)
			writeResponse(req, res)
		}
		return nil
	}

	if cmd.Name == DISCARD {
		if !req.IsInTxn() {
			return errors.New("DISCARD without MULTI")
		}

		req.DiscardTransaction()
		writeResponse(req, resp.NewString("OK"))
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
		res := resp.NewString("QUEUED")
		writeResponse(req, res)
		return nil
	}

	if req.SubMode && !spec.allowedInSubMode {
		return fmt.Errorf("Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context", cmdName)
	}

	err := spec.handler(req, cmd.Args)
	if err != nil {
		return err
	}

	if spec.cmdType == cmdTypeWrite && !isReplica {
		replicaCommand := utils.BulkStringsToRESPArray(append([]string{cmdName}, cmd.Args...))
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
