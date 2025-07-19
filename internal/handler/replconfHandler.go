package handler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils"
)

func replconfHandler(req *request.Request, args []string) error {
	if len(args) == 0 {
		return errors.New("REPLCONF requires at least one argument")
	}

	subcommand := strings.ToUpper(args[0])

	switch subcommand {
	case "GETACK":
		return replconfGETACK(req, args[1:])
	case "ACK":
		return replconfACK(req, args[1:])
	default:
		return writeResponse(req, resp.NewRESPString("OK"))
	}
}

func replconfGETACK(req *request.Request, args []string) error {
	if len(args) != 1 {
		return errors.New("REPLCONF GETACK requires exactly one argument")
	}

	offset := 0
	req.State.ReadState(func(s state.State) {
		offset = s.ReplicationOffset
	})

	command := utils.BulkStringsToRESPArray([]string{"REPLCONF", "ACK", strconv.Itoa(offset)})
	return writeResponse(req, command)
}

func replconfACK(req *request.Request, args []string) error {
	if len(args) != 1 {
		return errors.New("REPLCONF ACK requires exactly one argument")
	}
	offset, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid offset: %s", args[0])
	}

	replica, ok := req.State.GetReplica(req.Client.ID)
	if !ok {
		return fmt.Errorf("client is not a replica, %s", req.Client.RemoteAddr().String())
	}

	replica.Offset = offset

	fmt.Printf("Replica %s acknowledged offset %d\n", req.Client.ID, offset)

	select {
	case <-replica.OffsetChan:
	default:
	}

	replica.OffsetChan <- offset

	return nil
}
