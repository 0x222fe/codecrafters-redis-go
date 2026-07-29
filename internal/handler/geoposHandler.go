package handler

import (
	"errors"
	"fmt"

	"github.com/0x222fe/codecrafters-redis-go/internal/client"
	"github.com/0x222fe/codecrafters-redis-go/internal/state"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/geoutil"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
)

func geoposHandler(c *client.Client, s *state.AppState, args []string) error {
	if len(args) < 2 {
		return errors.New("GEOPOS requires at least 2 arguments")
	}

	key, locations := args[0], args[1:]

	arr := make([]resp.RESPValue, 0, len(locations))

	for _, location := range locations {
		score, ok := s.GetStore().QuerySortedSetScore(key, location)

		val := resp.RESPNilArray
		if ok {
			lo, la := geoutil.DecodeScore(score)
			val = resputil.BulkStringsToRESPArray([]string{fmt.Sprintf("%.17g", lo), fmt.Sprintf("%.17g", la)})
		}

		arr = append(arr, val)
	}

	res := resp.NewArray(arr)
	writeResponse(c, res)
	return nil
}
