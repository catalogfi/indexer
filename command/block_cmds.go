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
	height, err := l.store.GetLatestBlockHeight()
	if err != nil {
		if err.Error() == store.ErrKeyNotFound {
			return nil, store.ErrGetLatestBlockHeightNone
		}
		return nil, err
	}
	return height, nil
}

func LatestBlock(store *store.Storage) Command {
	return &latestBlock{
		store: store,
	}
}
