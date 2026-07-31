package handler

import (
	"errors"
	"strconv"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func infoHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) == 0 {
		return errors.New("INFO requires at least one argument")
	}

	if args[0] != "replication" {
		return errors.New("only 'replication' section is supported")
	}

	isReplica, repID, repOffset := false, "", 0
	s.ReadState(func(st state.ReplicaState) {
		isReplica = st.IsReplica
		repID = st.ReplicationID
		repOffset = st.ReplicationOffset
	})

	var role string
	if isReplica {
		role = "slave"
	} else {
		role = "master"
	}

	info := "# Replication\r\n" +
		"role:" + role + "\r\n"
	if !isReplica {
		info += "master_replid:" + repID + "\r\n" +
			"master_repl_offset:" + strconv.Itoa(repOffset) + "\r\n"
	}

	res := resp.NewBulkString(&info)

	return writeResponse(c, res)
}
