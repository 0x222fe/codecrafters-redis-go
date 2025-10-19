package handler

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func psyncHandler(req *request.Request, args []string) error {
	if len(args) < 2 {
		return errors.New("PSYNC requires at least 2 arguments")
	}

	if args[0] != "?" || args[1] != "-1" {
		return errors.New("PSYNC only supports ? -1 for now")
	}

	replicationID, replicationOffset := "", 0

	req.State.ReadState(func(s state.State) {
		replicationID = s.ReplicationID
		replicationOffset = s.ReplicationOffset
	})

	psyncRes := resp.NewString("FULLRESYNC " + replicationID + " " + strconv.Itoa(replicationOffset))

	err := writeResponse(req, psyncRes)
	if err != nil {
		return err
	}

	req.State.AddReplica(req.Client)

	emptyRdb := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="

	fileBytes, err := base64.StdEncoding.DecodeString(emptyRdb)
	if err != nil {
		return errors.New("failed to decode RDB file: " + err.Error())
	}

	header := fmt.Appendf(nil, "$%d\r\n", len(fileBytes))
	if _, err := req.Client.Write(header); err != nil {
		return fmt.Errorf("failed to write RDB header: %w", err)
	}

	if _, err := req.Client.Write(fileBytes); err != nil {
		return fmt.Errorf("failed to write RDB file: %w", err)
	}

	return nil
}
