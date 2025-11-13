package handler

import (
	"errors"
	"fmt"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
)

type commandSpec struct {
	handler          commandHandler
	cmdType          request.CommandType
	allowedInSubMode bool
}
type commandHandler func(req *request.Request, args []string) error

func (h commandHandler) Handle(req *request.Request, cmd request.Command) error {
	return h(req, cmd.Args)
}

var (
	handlerReg = map[request.CommandKey]commandSpec{
		request.PING:        {handler: pingHandler, allowedInSubMode: true},
		request.ECHO:        {handler: echoHandler},
		request.SET:         {handler: setHandler, cmdType: request.CmdTypeWrite},
		request.GET:         {handler: getHandler},
		request.CONFIG:      {handler: configHandler},
		request.KEYS:        {handler: keysHandler},
		request.INFO:        {handler: infoHandler},
		request.REPLCONF:    {handler: replconfHandler},
		request.PSYNC:       {handler: psyncHandler},
		request.WAIT:        {handler: waitHandler},
		request.TYPE:        {handler: typeHandler},
		request.XADD:        {handler: xaddHandler, cmdType: request.CmdTypeWrite},
		request.XRANGE:      {handler: xrangeHandler},
		request.XREAD:       {handler: xreadHandler},
		request.INCR:        {handler: incrHandler, cmdType: request.CmdTypeWrite},
		request.MULTI:       {handler: multiHandler},
		request.LPUSH:       {handler: lpushHandler, cmdType: request.CmdTypeWrite},
		request.RPUSH:       {handler: rpushHandler, cmdType: request.CmdTypeWrite},
		request.LRANGE:      {handler: lrangeHandler},
		request.LLEN:        {handler: llenHandler},
		request.LPOP:        {handler: lpopHandler, cmdType: request.CmdTypeWrite},
		request.BLPOP:       {handler: blpopHandler, cmdType: request.CmdTypeWrite},
		request.RPOP:        {handler: rpopHandler, cmdType: request.CmdTypeWrite},
		request.SUBSCRIBE:   {handler: subscribeHandler, allowedInSubMode: true},
		request.UNSUBSCRIBE: {handler: unsubscribeHandler, allowedInSubMode: true},
		request.PUBLISH:     {handler: publishHandler, cmdType: request.CmdTypeWrite, allowedInSubMode: true},
		request.ZADD:        {handler: zaddHandler, cmdType: request.CmdTypeWrite},
		request.ZRANK:       {handler: zrankHandler, cmdType: request.CmdTypeRead},
		request.ZRANGE:      {handler: zrangeHandler, cmdType: request.CmdTypeRead},
		request.ZCARD:       {handler: zcardHandler, cmdType: request.CmdTypeRead},
		request.ZSCORE:      {handler: zscoreHandler, cmdType: request.CmdTypeRead},
		request.ZREM:        {handler: zremHandler, cmdType: request.CmdTypeWrite},
		request.GEOADD:      {handler: geoaddHandler, cmdType: request.CmdTypeWrite},
		request.GEOPOS:      {handler: geoposHandler, cmdType: request.CmdTypeRead},
		request.GEODIST:     {handler: geodistHandler, cmdType: request.CmdTypeRead},
		request.GEOSEARCH:   {handler: geosearchHandler, cmdType: request.CmdTypeRead},
		request.ACL:         {handler: aclHandler, cmdType: request.CmdTypeRead},
	}
)

func RunCommand(req *request.Request, cmd request.Command) error {
	cmdName := string(cmd.Name)

	if cmd.Name == request.EXEC {
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

	if cmd.Name == request.DISCARD {
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

	if spec.cmdType == request.CmdTypeWrite &&
		isReplica &&
		!req.Propagated {
		return errors.New("replica cannot execute write commands")
	}

	if req.IsInTxn() {
		if cmd.Name == request.MULTI {
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

	if spec.cmdType == request.CmdTypeWrite && !isReplica {
		replicaCommand := resputil.BulkStringsToRESPArray(append([]string{cmdName}, cmd.Args...))
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
