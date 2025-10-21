package handler

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
)

func blpopHandler(req *request.Request, args []string) error {
	if len(args) < 2 {
		return errors.New("BLPOP requires at least 2 arguments")
	}

	keys, timeoutStr := args[:len(args)-1], args[len(args)-1]

	timeoutSec, err := strconv.ParseFloat(timeoutStr, 10)
	if err != nil {
		return fmt.Errorf("invalid timeout: %w", err)
	}

	if timeoutSec < 0 {
		return fmt.Errorf("timeout is not a positive number or zero")
	}

	for _, key := range keys {
		v, _, ok := req.State.GetStore().Get(key)
		list, parseOk := v.(*store.RedisList)
		if ok && !parseOk {
			return store.ERRWrongType
		}

		if !ok {
			continue
		}

		items, ok := list.LPop(1)
		if !ok {
			continue
		}
		writeResponse(
			req,
			resputil.BulkStringsToRESPArray([]string{key, items[0]}),
		)
		return nil
	}

	doneChan := make(chan [2]string, 1)
	var timeoutCh <-chan time.Time
	if timeoutSec > 0 {
		d := time.Duration(timeoutSec * float64(time.Second))
		timeoutCh = time.After(d)
	}

	defer func() {
		for _, key := range keys {
			req.State.GetStore().UnregisterListPushHandler(key, req.Client.ID)
		}
	}()

	for _, key := range keys {
		v, _, ok := req.State.GetStore().Get(key)
		_, parseOk := v.(*store.RedisList)
		if ok && !parseOk {
			return store.ERRWrongType
		}

		ch := make(chan string, 1)

		go func() {
			req.State.GetStore().RegisterListPushHandler(key, req.Client.ID, ch)
			item := <-ch
			doneChan <- [2]string{key, item}

		}()
	}

	select {
	case data := <-doneChan:
		key := data[0]
		v, _ := req.State.GetStore().GetExact(key, store.List)
		list, ok := v.(*store.RedisList)
		if ok {
			list.LPop(1)
		}
		writeResponse(req, resputil.BulkStringsToRESPArray(data[:]))
		return nil
	case <-timeoutCh:
		writeResponse(req, resp.RESPNilArray)
		return nil
	}
}
