package handler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
)

func replconfHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) == 0 {
		return errors.New("REPLCONF requires at least one argument")
	}

	subcommand := strings.ToUpper(args[0])

	switch subcommand {
	case "GETACK":
		return replconfGETACK(c, s, args[1:])
	case "ACK":
		return replconfACK(c, s, args[1:])
	default:
		return writeResponse(c, resp.NewString("OK"))
	}
}

func replconfGETACK(c *client.Client, s *state.AppState, args []string) error {
	if len(args) != 1 {
		return errors.New("REPLCONF GETACK requires exactly one argument")
	}

	offset := 0
	s.ReadState(func(st state.State) {
		offset = st.ReplicationOffset
	})

	command := resputil.BulkStringsToRESPArray([]string{"REPLCONF", "ACK", strconv.Itoa(offset)})
	return writeResponse(c, command)
}

func replconfACK(c *client.Client, s *state.AppState, args []string) error {
	if len(args) != 1 {
		return errors.New("REPLCONF ACK requires exactly one argument")
	}
	offset, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid offset: %s", args[0])
	}

	replica, ok := s.GetReplica(c.Conn.ID)
	if !ok {
		return fmt.Errorf("client is not a replica, %s", c.Conn.RemoteAddr().String())
	}

	replica.Offset = offset

	fmt.Printf("Replica %s acknowledged offset %d\n", c.Conn.ID, offset)

	for {
		select {
		case replica.OffsetChan <- offset:
			return nil
		case <-replica.OffsetChan:
			// Channel full, drain and retry
		case <-replica.Ctx.Done():
			return errors.New("replica context canceled")
		}
	}
}
