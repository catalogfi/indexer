package command

import (
	"encoding/json"

	"github.com/catalogfi/indexer/store"
)

type latestTip struct {
	store *store.Storage
}

func (l *latestTip) Name() string {
	return "latest_tip"
}

func (l *latestTip) Execute(params json.RawMessage) (interface{}, error) {
	height, exists, err := l.store.GetLatestBlockHeight()
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, store.ErrGetLatestBlockHeightNone
	}
	return height, nil
}

func LatestTip(store *store.Storage) Command {
	return &latestTip{
		store: store,
	}
}

// latestTipHash

type latestTipHash struct {
	store *store.Storage
}

func (l *latestTipHash) Name() string {
	return "latest_tip_hash"
}

func (l *latestTipHash) Execute(params json.RawMessage) (interface{}, error) {
	hash, exists, err := l.store.GetLatestTipHash()
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, store.ErrGetLatestTipHash
	}
	return hash, nil
}

func LatestTipHash(store *store.Storage) Command {
	return &latestTipHash{
		store: store,
	}
}
