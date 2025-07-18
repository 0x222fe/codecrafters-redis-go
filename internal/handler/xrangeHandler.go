package handler

import (
	"errors"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils"
)

func xrangeHandler(req *request.Request, args []string) error {
	if len(args) != 3 {
		return errors.New("XRANGE requires 3 arguments")
	}
	key, startS, endS := args[0], args[1], args[2]

	var start, end []byte = nil, nil
	if startS != "-" {
		s, err := store.ParseStreamEntryID(startS)
		if err != nil {
			return errors.New("invalid start")
		}
		start = s.RadixKey()
	}

	if endS != "+" {
		e, err := store.ParseStreamEntryID(endS)
		if err != nil {
			return errors.New("invalid end")
		}
		end = e.RadixKey()
	}

	v, ok := req.State.GetStore().GetExact(key, store.Stream)
	stream, parseOk := v.(*store.RedisStream)
	if !ok || !parseOk {
		return errors.New("key is not a stream")
	}

	var res resp.RESPValue
	if ok {
		entries := stream.Range(start, end)
		res = utils.StreamEntriesToRESPArray(entries)
	} else {
		res = utils.StreamEntriesToRESPArray(nil)
	}

	writeResponse(req, res)
	return nil
}
