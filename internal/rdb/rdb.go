package rdb

type RDB struct {
	Version   string
	Metadata  map[string]string
	Databases map[int]*Database
}

type Database struct {
	Index               int
	HashTableSize       int
	ExpiryHashTableSize int
	Items               map[string]*KeyValue
}

type KeyValue struct {
	Key      string
	Value    string
	Type     byte
	ExpireAt *int64
}

const (
	metaFlag  = 0xFA
	tableFlag = 0xFB
	msExpFlag = 0xFC
	sExpFlag  = 0xFD
	dbFlag    = 0xFE
	endFlag   = 0xFF
)
