package handler

import (
	"errors"
	"fmt"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/command"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

type commandSpec struct {
	handler          commandHandler
	cmdType          command.CommandType
	allowedInSubMode bool
}
type commandHandler func(c *client.Client, s *state.AppState, args []string) error

func (h commandHandler) Handle(c *client.Client, s *state.AppState, cmd command.Command) error {
	return h(c, s, cmd.Args)
}

var (
	handlerReg = map[command.CommandKey]commandSpec{
		command.PING:        {handler: pingHandler, allowedInSubMode: true},
		command.ECHO:        {handler: echoHandler},
		command.SET:         {handler: setHandler, cmdType: command.TypeWrite},
		command.GET:         {handler: getHandler},
		command.CONFIG:      {handler: configHandler},
		command.KEYS:        {handler: keysHandler},
		command.INFO:        {handler: infoHandler},
		command.REPLCONF:    {handler: replconfHandler},
		command.PSYNC:       {handler: psyncHandler},
		command.WAIT:        {handler: waitHandler},
		command.TYPE:        {handler: typeHandler},
		command.XADD:        {handler: xaddHandler, cmdType: command.TypeWrite},
		command.XRANGE:      {handler: xrangeHandler},
		command.XREAD:       {handler: xreadHandler},
		command.INCR:        {handler: incrHandler, cmdType: command.TypeWrite},
		command.MULTI:       {handler: multiHandler},
		command.LPUSH:       {handler: lpushHandler, cmdType: command.TypeWrite},
		command.RPUSH:       {handler: rpushHandler, cmdType: command.TypeWrite},
		command.LRANGE:      {handler: lrangeHandler},
		command.LLEN:        {handler: llenHandler},
		command.LPOP:        {handler: lpopHandler, cmdType: command.TypeWrite},
		command.BLPOP:       {handler: blpopHandler, cmdType: command.TypeWrite},
		command.RPOP:        {handler: rpopHandler, cmdType: command.TypeWrite},
		command.SUBSCRIBE:   {handler: subscribeHandler, allowedInSubMode: true},
		command.UNSUBSCRIBE: {handler: unsubscribeHandler, allowedInSubMode: true},
		command.PUBLISH:     {handler: publishHandler, cmdType: command.TypeWrite, allowedInSubMode: true},
		command.ZADD:        {handler: zaddHandler, cmdType: command.TypeWrite},
		command.ZRANK:       {handler: zrankHandler, cmdType: command.TypeRead},
		command.ZRANGE:      {handler: zrangeHandler, cmdType: command.TypeRead},
		command.ZCARD:       {handler: zcardHandler, cmdType: command.TypeRead},
		command.ZSCORE:      {handler: zscoreHandler, cmdType: command.TypeRead},
		command.ZREM:        {handler: zremHandler, cmdType: command.TypeWrite},
		command.GEOADD:      {handler: geoaddHandler, cmdType: command.TypeWrite},
		command.GEOPOS:      {handler: geoposHandler, cmdType: command.TypeRead},
		command.GEODIST:     {handler: geodistHandler, cmdType: command.TypeRead},
		command.GEOSEARCH:   {handler: geosearchHandler, cmdType: command.TypeRead},
		command.ACL:         {handler: aclHandler, cmdType: command.TypeRead},
		command.AUTH:        {handler: authHandler, cmdType: command.TypeRead},
		command.WATCH:       {handler: watchHandler, cmdType: command.TypeRead},
		command.UNWATCH:     {handler: unwatchHandler, cmdType: command.TypeRead},
	}
)

func RunCommand(c *client.Client, s *state.AppState, cmd command.Command) error {
	cmdName := string(cmd.Name)

	if cmd.Name == command.EXEC {
		if !c.IsInTxn() {
			return errors.New("EXEC without MULTI")
		}

		resArr, err := c.ExecTransaction(s)
		if err != nil {
			return fmt.Errorf("failed to execute transaction: %w", err)
		}

		res := resp.NewArray(resArr)
		writeResponse(c, res)
		return nil
	}

	if cmd.Name == command.DISCARD {
		if !c.IsInTxn() {
			return errors.New("DISCARD without MULTI")
		}

		c.DiscardTransaction(s)
		writeResponse(c, resp.NewString("OK"))
		return nil
	}

	spec, find := handlerReg[cmd.Name]
	if !find {
		return errors.New("unknown command: " + cmdName)
	}

	var isReplica bool
	s.ReadState(func(st state.ReplicaState) {
		isReplica = st.IsReplica
	})

	if spec.cmdType == command.TypeWrite &&
		isReplica &&
		!c.Propagated {
		return errors.New("replica cannot execute write commands")
	}

	if c.IsInTxn() {
		if cmd.Name == command.MULTI {
			return errors.New("MULTI calls can not be nested")
		}

		if cmd.Name == command.WATCH || cmd.Name == command.UNWATCH {
			return fmt.Errorf("%s inside MULTI is not allowed", cmd.Name)
		}

		txnCmds := c.Transaction.Commands
		txnCmds = append(txnCmds, client.TxnCommand{Command: cmd, Handler: spec.handler})
		c.Transaction.Commands = txnCmds
		res := resp.NewString("QUEUED")
		writeResponse(c, res)
		return nil
	}

	if c.SubMode && !spec.allowedInSubMode {
		return fmt.Errorf("Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context", cmdName)
	}

	err := spec.handler(c, s, cmd.Args)
	if err != nil {
		return err
	}

	if spec.cmdType == command.TypeWrite && !isReplica {
		replicaCommand := cmd.EncodeRESP()
		encoded := replicaCommand.Bytes()

		s.WriteState(func(st *state.ReplicaState) {
			st.ReplicationOffset += len(encoded)
		})

		replicas := s.GetReplicas()

		for _, rep := range replicas {
			if _, err := rep.Conn.Write(encoded); err != nil {
				fmt.Printf("failed to propagate command to replica %s: %v\n", rep.Conn.RemoteAddr(), err)
			}
		}
	}

	return nil
}

func writeResponse(c *client.Client, res resp.RESPValue) error {
	writer := c.GetWriter()
	err := writer.WriteResp(res)
	if err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	fmt.Printf("Response sent to %s\n", c.Conn.RemoteAddr())
	return nil
}
