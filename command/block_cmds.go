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
	height, exists, err := l.store.GetLatestBlockHeight()
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, store.ErrGetLatestBlockHeightNone
	}
	return height, nil
}

func LatestBlock(store *store.Storage) Command {
	return &latestBlock{
		store: store,
	}
}
