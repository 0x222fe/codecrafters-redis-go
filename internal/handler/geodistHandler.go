package handler

import (
	"errors"
	"fmt"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/geoutil"
)

func geodistHandler(req *request.Request, args []string) error {
	if len(args) != 3 {
		return errors.New("GEODIST requires exactly 3 arguments")
	}

	key, a, b := args[0], args[1], args[2]

	aScore, ok := req.State.GetStore().QuerySortedSetScore(key, a)
	if !ok {
		return fmt.Errorf("GEODIST: member %s not found", a)
	}

	bScore, ok := req.State.GetStore().QuerySortedSetScore(key, b)
	if !ok {
		return fmt.Errorf("GEODIST: member %s not found", b)
	}

	loa, laa := geoutil.DecodeScore(aScore)
	lob, lab := geoutil.DecodeScore(bScore)

	dist := geoutil.Distance(loa, laa, lob, lab)

	str := fmt.Sprintf("%.4f", dist)

	res := resp.NewBulkString(&str)
	writeResponse(req, res)
	return nil
}