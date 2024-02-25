package store

import "errors"

var (
	ErrKeyNotFound = "key not found"
)

// Storage errors
var (
	ErrGetLatestBlockHeightNone = errors.New("latest block height not found. Did you forget to run the indexer?")
	ErrGetLatestTipHash         = errors.New("latest tip hash not found")
	ErrGetTxNotFound            = errors.New("transaction not found")
)
