package command

import (
	"errors"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/internal/state"
)

func infoHandler(state *state.AppState, args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("INFO requires at least one argument")
	}

	if args[0] != "replication" {
		return nil, errors.New("only 'replication' section is supported")
	}

	var role string
	if state.Cfg.ReplicaHost != "" && state.Cfg.ReplicaPort != 0 {
		role = "slave"
	} else {
		role = "master"
	}

	info := "# Replication\r\n" +
		"role:" + role + "\r\n"

	result := fmt.Sprintf("$%d\r\n%s\r\n", len(info), info)

	return []byte(result), nil
}
