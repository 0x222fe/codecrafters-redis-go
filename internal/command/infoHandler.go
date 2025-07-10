package command

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func infoHandler(state *state.AppState, args []string, writer io.Writer) error {
	if len(args) == 0 {
		return errors.New("INFO requires at least one argument")
	}

	if args[0] != "replication" {
		return errors.New("only 'replication' section is supported")
	}

	var role string
	if state.IsReplica {
		role = "slave"
	} else {
		role = "master"
	}

	info := "# Replication\r\n" +
		"role:" + role + "\r\n"
	if !state.IsReplica {
		info += "master_replid:" + state.ReplicationID + "\r\n" +
			"master_repl_offset:" + strconv.Itoa(state.ReplicationOffset) + "\r\n"
	}

	result := fmt.Sprintf("$%d\r\n%s\r\n", len(info), info)

	return writeResponse(writer, []byte(result))
}
