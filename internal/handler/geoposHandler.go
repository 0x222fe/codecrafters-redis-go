package handler

import (
	"errors"
	"fmt"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/geoutil"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
)

func geoposHandler(req *request.Request, args []string) error {
	if len(args) < 2 {
		return errors.New("GEOADD requires at least 2 arguments")
	}

	key, locations := args[0], args[1:]

	arr := make([]resp.RESPValue, 0, len(locations))

	for _, location := range locations {
		score, ok := req.State.GetStore().QuerySortedSetScore(key, location)

		val := resp.RESPNilArray
		if ok {
			lo, la := geoutil.DecodeScore(score)
			val = resputil.BulkStringsToRESPArray([]string{fmt.Sprintf("%.17g", lo), fmt.Sprintf("%.17g", la)})
		}

		arr = append(arr, val)
	}

	res := resp.NewArray(arr)
	writeResponse(req, res)
	return nil
}
