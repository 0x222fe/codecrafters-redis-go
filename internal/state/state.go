package state

import (
	"github.com/0x222fe/codecrafters-redis-go/internal/config"
	"github.com/0x222fe/codecrafters-redis-go/internal/store"
)

type AppState struct {
	Cfg                *config.Config
	Store              *store.Store
	IsReplica          bool
	ReplicantionID     string
	ReplicantionOffset int
}
