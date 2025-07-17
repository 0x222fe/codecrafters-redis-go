package store

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	iradix "github.com/hashicorp/go-immutable-radix/v2"
)

type RedisStream struct {
	mu   sync.RWMutex
	tree *iradix.Tree[*StreamEntry]
}

type StreamEntryID struct {
	millis   uint64
	sequence uint64
}

type StreamEntry struct {
	ID     StreamEntryID
	Fields map[string]string
}

func NewStream(key string) *RedisStream {
	return &RedisStream{
		tree: iradix.New[*StreamEntry](),
	}
}

func (id StreamEntryID) RadixKey() []byte {
	var b [16]byte
	binary.BigEndian.PutUint64(b[:8], uint64(id.millis))
	binary.BigEndian.PutUint64(b[8:], uint64(id.sequence))
	return b[:]
}

func (id StreamEntryID) String() string {
	return fmt.Sprintf("%d-%d", id.millis, id.sequence)
}

func (stream *RedisStream) GetItem(idStr string) (*StreamEntry, bool) {
	id, err := ParseStreamEntryID(idStr)
	if err != nil {
		return nil, false
	}
	key := id.RadixKey()

	stream.mu.RLock()
	defer stream.mu.RUnlock()
	entry, ok := stream.tree.Get(key)
	if !ok {
		return nil, false
	}

	return entry, true
}

func (stream *RedisStream) AddEntry(idStr string, fields map[string]string) (StreamEntryID, error) {
	millisP, seqP, err := validateStreamEntryIDInput(idStr)
	if err != nil {
		return StreamEntryID{}, fmt.Errorf("invalid stream entry ID: %s", err)
	}

	stream.mu.Lock()
	defer stream.mu.Unlock()

	var millis, seq uint64
	_, top, ok := stream.tree.Root().Maximum()
	switch {
	case millisP == nil && seqP == nil:
		if ok {
			millis = top.ID.millis
			seq = top.ID.sequence + 1
		} else {
			millis = uint64(time.Now().UnixMilli())
			seq = 0
		}
	case millisP != nil && seqP == nil:
		millis = *millisP
		if ok {
			if millis < top.ID.millis {
				return StreamEntryID{}, errors.New("The ID specified in XADD is equal or smaller than the target stream top item")
			}

			if millis == top.ID.millis {
				seq = top.ID.sequence + 1
			} else {
				seq = 0
			}
		} else {
			if millis == 0 {
				seq = 1
			} else {
				seq = 0
			}
		}
	case millisP != nil && seqP != nil:
		millis = *millisP
		seq = *seqP
		if millis == 0 && seq == 0 {
			return StreamEntryID{}, errors.New("The ID specified in XADD must be greater than 0-0")
		}

		if ok {
			if millis < top.ID.millis || (millis == top.ID.millis && seq <= top.ID.sequence) {
				return StreamEntryID{}, errors.New("The ID specified in XADD is equal or smaller than the target stream top item")
			}
		}
	default:
		return StreamEntryID{}, errors.New("invalid ID input")
	}

	key := make([]byte, 16)
	binary.BigEndian.PutUint64(key[:8], millis)
	binary.BigEndian.PutUint64(key[8:], seq)

	entry := &StreamEntry{
		ID:     StreamEntryID{millis: millis, sequence: seq},
		Fields: fields,
	}

	tree, _, _ := stream.tree.Insert(key, entry)
	stream.tree = tree

	return entry.ID, nil
}

func (stream *RedisStream) Range(startKey, endKey []byte) []*StreamEntry {
	stream.mu.RLock()
	defer stream.mu.RUnlock()

	result := make([]*StreamEntry, 0)
	it := stream.tree.Root().Iterator()
	it.SeekLowerBound(startKey)
	for {
		key, entry, ok := it.Next()
		if !ok {
			break
		}
		if bytes.Compare(key, endKey) > 0 {
			break
		}
		result = append(result, entry)
	}
	return result
}

func ParseStreamEntryID(str string) (StreamEntryID, error) {
	var millis, sequence uint64
	n, err := fmt.Sscanf(str, "%d-%d", &millis, &sequence)
	if n != 2 || err != nil {
		return StreamEntryID{}, fmt.Errorf("invalid stream entry ID format: %s", str)
	}
	return StreamEntryID{millis: millis, sequence: sequence}, nil
}

func validateStreamEntryIDInput(id string) (*uint64, *uint64, error) {
	id = strings.TrimSpace(id)
	if id == "*" {
		return nil, nil, nil
	}

	parts := strings.Split(id, "-")
	if len(parts) == 2 {
		millis, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid millis part: %w", err)
		}
		if parts[1] == "*" {
			return &millis, nil, nil
		}
		seq, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid sequence part: %w", err)
		}
		return &millis, &seq, nil
	}

	return nil, nil, errors.New("invalid format")
}
