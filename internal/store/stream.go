package store

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-immutable-radix/v2"
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
	id     StreamEntryID
	fields map[string]string
}

func NewStream(key string) *RedisStream {
	return &RedisStream{
		tree: iradix.New[*StreamEntry](),
	}
}

func (id StreamEntryID) radixKey() []byte {
	var b [16]byte
	binary.BigEndian.PutUint64(b[:8], uint64(id.millis))
	binary.BigEndian.PutUint64(b[8:], uint64(id.sequence))
	return b[:]
}

func (id StreamEntryID) String() string {
	return fmt.Sprintf("%d-%d", id.millis, id.sequence)
}

func (stream *RedisStream) GetItem(idStr string) (*StreamEntry, bool) {
	id, err := parseStreamEntryID(idStr)
	if err != nil {
		return nil, false
	}
	key := id.radixKey()

	stream.mu.RLock()
	defer stream.mu.RUnlock()
	entry, ok := stream.tree.Get(key)
	if !ok {
		return nil, false
	}

	return entry, true
}

func (stream *RedisStream) AddEntry(idStr string, fields map[string]string) (StreamEntryID, error) {
	if idStr != "*" {
		id, err := parseStreamEntryID(idStr)
		if err != nil {
			return StreamEntryID{}, fmt.Errorf("invalid stream entry ID: %s", err)
		}
		key := id.radixKey()

		stream.mu.Lock()
		defer stream.mu.Unlock()
		_, ok := stream.tree.Get(key)

		if ok {
			return StreamEntryID{}, fmt.Errorf("stream entry already exists: %s", idStr)
		}

		entry := &StreamEntry{
			id:     id,
			fields: fields,
		}
		t, _, _ := stream.tree.Insert(key, entry)
		stream.tree = t
		return id, nil
	}

	millis := uint64(time.Now().UnixMilli())
	key := make([]byte, 16)
	binary.BigEndian.PutUint64(key[:8], millis)

	maxSeq := uint64(0)
	stream.mu.Lock()
	defer stream.mu.Unlock()
	stream.tree.Root().WalkPrefix(key, func(k []byte, v *StreamEntry) bool {
		candidate := binary.BigEndian.Uint64(k[8:])
		if candidate > maxSeq {
			maxSeq = candidate
		}
		return false
	})
	seq := maxSeq + 1
	binary.BigEndian.PutUint64(key[8:], seq)

	entry := &StreamEntry{
		id: StreamEntryID{
			millis:   millis,
			sequence: seq,
		},
		fields: fields,
	}

	tree, _, ok := stream.tree.Insert(key, entry)
	if !ok {
		return StreamEntryID{}, fmt.Errorf("insert failed")
	}

	stream.tree = tree

	return StreamEntryID{millis: millis, sequence: seq}, nil
}

func parseStreamEntryID(str string) (StreamEntryID, error) {
	var millis, sequence uint64
	n, err := fmt.Sscanf(str, "%d-%d", &millis, &sequence)
	if n != 2 || err != nil {
		return StreamEntryID{}, fmt.Errorf("invalid stream entry ID format: %s", str)
	}
	return StreamEntryID{millis: millis, sequence: sequence}, nil
}
