package command

import (
	"encoding/json"

	"github.com/catalogfi/indexer/store"
)

type latestBlock struct {
	store *store.Storage
}

func (l *latestBlock) Name() string {
	return "latestblock"
}

func (l *latestBlock) Execute(params json.RawMessage) (interface{}, error) {
	return l.store.GetLatestBlockHeight()
}

func LatestBlock(store *store.Storage) Command {
	return &latestBlock{
		store: store,
	}
}
