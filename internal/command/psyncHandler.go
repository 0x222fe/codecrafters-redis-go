package command

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/0x222fe/codecrafters-redis-go/internal/cnn"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
)

func psyncHandler(appState *state.AppState, args []string, conn *cnn.Connection) error {
	if len(args) < 2 {
		return errors.New("PSYNC requires at least 2 arguments")
	}

	if args[0] != "?" || args[1] != "-1" {
		return errors.New("PSYNC only supports ? -1 for now")
	}

	var replicationID string

	appState.ReadState(func(s state.State) {
		replicationID = s.ReplicationID
	})

	psyncMsg := resp.NewRESPString("FULLRESYNC " + replicationID + " " + "0").Encode()

	err := writeResponse(conn, psyncMsg)
	if err != nil {
		return err
	}

	appState.AddReplica(conn)

	emptyRdb := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="

	fileBytes, err := base64.StdEncoding.DecodeString(emptyRdb)
	if err != nil {
		return errors.New("failed to decode RDB file: " + err.Error())
	}

	header := fmt.Appendf(nil, "$%d\r\n", len(fileBytes))
	if _, err := conn.Write(header); err != nil {
		return fmt.Errorf("failed to write RDB header: %w", err)
	}

	if _, err := conn.Write(fileBytes); err != nil {
		return fmt.Errorf("failed to write RDB file: %w", err)
	}

	return nil
}
