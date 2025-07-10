package state

import (
	"github.com/codecrafters-io/redis-starter-go/internal/config"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

type AppState struct {
	Cfg                *config.Config
	Store              *store.Store
	IsReplica          bool
	ReplicantionID     string
	ReplicantionOffset int
}
