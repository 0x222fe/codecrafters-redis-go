package command

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func infoHandler(appState *state.AppState, args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("INFO requires at least one argument")
	}

	if args[0] != "replication" {
		return nil, errors.New("only 'replication' section is supported")
	}

	isReplica, repID, repOffset := false, "", 0
	appState.ReadState(func(s state.State) {
		isReplica = s.IsReplica
		repID = s.ReplicationID
		repOffset = s.ReplicationOffset
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

	result := fmt.Sprintf("$%d\r\n%s\r\n", len(info), info)

	return []byte(result), nil
}
