package mongodb

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/catalogfi/indexer/command"
	"github.com/catalogfi/indexer/peer"
	"go.mongodb.org/mongo-driver/mongo"
)

type Storage interface {
	command.Storage
	peer.Storage
}

type storage struct {
	params *chaincfg.Params
	db     *mongo.Database
}

func NewStorage(params *chaincfg.Params, db *mongo.Database) Storage {
	return &storage{
		params: params,
		db:     db,
	}
}
