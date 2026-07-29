package handler

import (
	"errors"
	"fmt"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/geoutil"
)

func geodistHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) != 3 {
		return errors.New("GEODIST requires exactly 3 arguments")
	}

	key, a, b := args[0], args[1], args[2]

	aScore, ok := s.GetStore().QuerySortedSetScore(key, a)
	if !ok {
		return fmt.Errorf("GEODIST: member %s not found", a)
	}

	bScore, ok := s.GetStore().QuerySortedSetScore(key, b)
	if !ok {
		return fmt.Errorf("GEODIST: member %s not found", b)
	}

	loa, laa := geoutil.DecodeScore(aScore)
	lob, lab := geoutil.DecodeScore(bScore)

	dist := geoutil.Distance(loa, laa, lob, lab)

	str := fmt.Sprintf("%.4f", dist)

	res := resp.NewBulkString(&str)
	writeResponse(c, res)
	return nil
}
