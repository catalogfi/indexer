package store

import (
	"github.com/catalogfi/indexer/database"
	"go.uber.org/zap"
)

// TODO: test reorgs
// TODO: test pending transactions

type Storage struct {
	db     database.Db
	logger *zap.Logger
}

func NewStorage(db database.Db) *Storage {
	logger := zap.NewNop()
	return &Storage{
		db:     db,
		logger: logger,
	}
}

func (s *Storage) SetLogger(logger *zap.Logger) *Storage {
	s.logger = logger
	return s
}
