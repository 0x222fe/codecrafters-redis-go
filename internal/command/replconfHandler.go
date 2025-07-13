package command

import (
	"errors"
	"strconv"
	"strings"

	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils"
)

func replconfHandler(state *state.AppState, args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("REPLCONF requires at least one argument")
	}

	subcommand := strings.ToUpper(args[0])

	switch subcommand {
	case "GETACK":
		return replconfGETACK(state, args[1:])
	case "ACK":
		return replconfACK(state, args[1:])
	default:
		// return nil, errors.New("Unknown REPLCONF subcommand: " + subcommand)
		return resp.NewRESPString("OK").Encode(), nil
	}
}

func replconfGETACK(appState *state.AppState, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("REPLCONF GETACK requires exactly one argument")
	}

	offset := 0
	appState.ReadState(func(s state.State) {
		offset = s.ReplicationOffset
	})

	return utils.EncodeStringSliceToRESP([]string{"REPLCONF", "ACK", strconv.Itoa(offset)}), nil
}

func replconfACK(state *state.AppState, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("REPLCONF ACK requires exactly one argument")
	}

	return nil, nil
}
