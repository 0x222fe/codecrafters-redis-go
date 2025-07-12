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
		return replconfGetAck(state, args[1:])
	// case "ACK":
	// 	return replconfAck(state, args[1:])
	default:
		// return nil, errors.New("Unknown REPLCONF subcommand: " + subcommand)
		return resp.NewRESPString("OK").Encode(), nil
	}
}

func replconfGetAck(_ *state.AppState, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("REPLCONF GETACK requires exactly one argument")
	}

	return utils.EncodeStringSliceToRESP([]string{"REPLCONF", "ACK", strconv.Itoa(0)}), nil
}

// func replconfAck(state *state.AppState, args []string) ([]byte, error) {
// 	if len(args) != 1 {
// 		return nil, errors.New("REPLCONF ACK requires exactly one argument")
// 	}
//
// 	return replconfGetAck(state, args)
// }
