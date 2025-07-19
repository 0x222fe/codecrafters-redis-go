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
	"github.com/0x222fe/codecrafters-redis-go/internal/utils"
	"github.com/google/uuid"
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
		type handler struct {
			key string
			id  uuid.UUID
		}
		type streamEntry struct {
			streamKey string
			entry     *store.StreamEntry
		}

		handlers := make([]handler, 0, len(keys))
		defer func() {
			for _, h := range handlers {
				req.State.GetStore().UnregisterStreamInsertHandler(h.key, h.id)
			}
		}()

		doneCh := make(chan streamEntry, 1)
		var timeoutCh <-chan time.Time

		if *blockMillis != 0 {
			ticker := time.NewTicker(time.Duration(*blockMillis) * time.Millisecond)
			defer ticker.Stop()
			timeoutCh = ticker.C
		}

		for _, key := range keys {
			localKey := key
			id := req.State.GetStore().RegisterStreamInsertHandler(key, func(entry *store.StreamEntry) {
				doneCh <- streamEntry{streamKey: localKey, entry: entry}
			})
			handlers = append(handlers, handler{key: key, id: id})
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
		res = resp.RESPNilBulkString
	} else {
		arr := make([]resp.RESPValue, 0)
		for _, key := range keys {
			entries := entryDict[key]
			if len(entries) == 0 {
				continue
			}
			streamArr := make([]resp.RESPValue, 0, 2*len(entries))
			streamArr = append(streamArr, resp.NewRESPBulkString(&key))
			streamEntryRESP := utils.StreamEntriesToRESPArray(entries)
			streamArr = append(streamArr, streamEntryRESP)
			arr = append(arr, resp.NewRESPArray(streamArr))
		}

		res = resp.NewRESPArray(arr)
	}

	writeResponse(req, res)
	return nil
}
