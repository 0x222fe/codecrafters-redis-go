package handler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/0x222fe/codecrafters-redis-go/internal/request"
	"github.com/0x222fe/codecrafters-redis-go/internal/resp"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
)

func xreadHandler(req *request.Request, args []string) error {
	if len(args) < 3 {
		return errors.New("XREAD requires at least 3 arguments")
	}

	var blockMillis *int
	if strings.ToUpper(args[0]) == "BLOCK" {
		t, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("error parsing BLOCK timeout: %w", err)
		}
		blockMillis = &t

		args = args[2:]
	}

	count := len(args) - 1
	if count%2 != 0 {
		return errors.New("XREAD must have even numer of arguments after 'streams'")
	}
	count /= 2
	keys, idStrs := args[1:1+count], args[1+count:]
	streams := make([]*store.RedisStream, 0, count)
	for _, key := range keys {
		v, _, has := req.State.GetStore().Get(key)
		stream, parseOk := v.(*store.RedisStream)
		if has && !parseOk {
			return store.ERRWrongType
		}
		streams = append(streams, stream)
	}

	idPtrs := make([]*store.StreamEntryID, 0, count)
	for i := range count {
		if idStrs[i] == "$" {
			idPtrs = append(idPtrs, nil)
			continue
		}

		id, err := store.ParseStreamEntryID(idStrs[i])
		if err != nil {
			return err
		}
		id.Seq++
		idPtrs = append(idPtrs, &id)
	}

	entryDict, fetchedCount := make(map[string][]*store.StreamEntry), 0
	for i, key := range keys {
		stream := streams[i]
		if stream == nil {
			continue
		}

		if idPtrs[i] != nil {
			entries := stream.Range((*idPtrs[i]).RadixKey(), nil)
			entryDict[key] = entries
			fetchedCount += len(entries)
		}

	}

	if fetchedCount == 0 && blockMillis != nil {
		type streamEntry struct {
			streamKey string
			entry     *store.StreamEntry
		}

		defer func() {
			for _, key := range keys {
				req.State.GetStore().UnregisterStreamInsertHandler(key, req.Client.ID)
			}
		}()

		doneCh := make(chan streamEntry, 1)
		var timeoutCh <-chan time.Time

		if *blockMillis != 0 {
			d := time.Duration(*blockMillis * int(time.Millisecond))
			timeoutCh = time.After(d)
		}

		for _, key := range keys {
			localKey := key
			req.State.GetStore().RegisterStreamInsertHandler(key, req.Client.ID, func(entry *store.StreamEntry) {
				doneCh <- streamEntry{streamKey: localKey, entry: entry}
			})
		}

		select {
		case d := <-doneCh:
			entryDict[d.streamKey] = append(entryDict[d.streamKey], d.entry)
			fetchedCount++
		case <-timeoutCh:
		}
	}

	var res resp.RESPValue
	if fetchedCount == 0 {
		res = resp.RESPNilArray
	} else {
		arr := make([]resp.RESPValue, 0)
		for _, key := range keys {
			entries := entryDict[key]
			if len(entries) == 0 {
				continue
			}
			streamArr := make([]resp.RESPValue, 0, 2*len(entries))
			streamArr = append(streamArr, resp.NewBulkString(&key))
			streamEntryRESP := resputil.StreamEntriesToRESPArray(entries)
			streamArr = append(streamArr, streamEntryRESP)
			arr = append(arr, resp.NewArray(streamArr))
		}

		res = resp.NewArray(arr)
	}

	writeResponse(req, res)
	return nil
}
