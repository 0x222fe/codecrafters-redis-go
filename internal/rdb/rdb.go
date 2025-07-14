package rdb

import (
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

type RDB struct {
	version   string
	metadata  map[string]string
	databases map[int]*database
}

type database struct {
	index               int
	hashTableSize       int
	expiryHashTableSize int
	items               map[string]*keyValue
}

type keyValue struct {
	key       string
	value     string
	valueType store.ValueType
	expireAt  *int64
}

const (
	metaFlag  = 0xFA
	tableFlag = 0xFB
	msExpFlag = 0xFC
	sExpFlag  = 0xFD
	dbFlag    = 0xFE
	endFlag   = 0xFF
)

func (rdb *RDB) MapToStore() *store.Store {
	s := store.NewStore()
	if rdb == nil {
		return s
	}

	for _, db := range rdb.databases {
		for _, kv := range db.items {
			s.Set(kv.key, kv.value, kv.valueType, kv.expireAt)
		}
	}
	return s
}
