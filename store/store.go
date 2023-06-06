package store

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/catalogfi/indexer/command"
	"github.com/catalogfi/indexer/peer"
	"gorm.io/gorm"
)

// TODO: test reorgs
// TODO: test pending transactions

type Storage interface {
	command.Storage
	peer.Storage
}

type storage struct {
	params *chaincfg.Params
	db     *gorm.DB
}

func NewStorage(params *chaincfg.Params, db *gorm.DB) Storage {
	return &storage{
		params: params,
		db:     db,
	}
}