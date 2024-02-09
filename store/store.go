package store

import (
	"github.com/catalogfi/indexer/database"
)

// TODO: test reorgs
// TODO: test pending transactions

type Storage struct {
	db database.Db
}

func NewStorage(db database.Db) *Storage {
	return &Storage{
		db: db,
	}
}
