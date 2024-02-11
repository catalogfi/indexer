package database

import (
	"github.com/erigontech/mdbx-go/mdbx"
	"go.uber.org/zap"
)

type MdbxDb struct {
	env    *mdbx.Env
	dbName string
	dbi    mdbx.DBI
	logger *zap.Logger
}

func NewMDBX(path string, dbName string) (*MdbxDb, error) {
	env, err := mdbx.NewEnv()
	if err != nil {
		return nil, err
	}

	err = env.SetOption(mdbx.OptMaxDB, 1)
	if err != nil {
		return nil, err
	}
	//TODO: optimize this and understand :/
	pageSize := mdbx.MaxPageSize
	err = env.SetGeometry(-1, -1, 1024*1024*pageSize, -1, -1, pageSize)
	if err != nil {
		return nil, err
	}
	//give all permissions to the file
	err = env.Open(path, 0, 0644)
	if err != nil {
		env.Close()
		return nil, err
	}
	logger := zap.NewNop()
	return (&MdbxDb{env: env, dbName: dbName, logger: logger}).OpenDbi()
}

func (m *MdbxDb) SetLogger(logger *zap.Logger) {
	m.logger = logger
}

func (m *MdbxDb) Close() {
	m.env.CloseDBI(m.dbi)
	m.env.Close()
}

func (m *MdbxDb) Get(key string) ([]byte, error) {
	var value []byte
	err := m.env.Update(func(txn *mdbx.Txn) error {
		var err error
		value, err = txn.Get(m.dbi, []byte(key))
		return err
	})
	return value, err
}

func (m *MdbxDb) Put(key string, value []byte) error {
	err := m.env.Update(func(txn *mdbx.Txn) error {
		return txn.Put(m.dbi, []byte(key), value, 0)
	})
	if err != nil {
		m.logger.Error("error putting value", zap.String("key", key), zap.Error(err))
	}
	return err
}

func (m *MdbxDb) Delete(key string) error {
	err := m.env.Update(func(txn *mdbx.Txn) error {
		return txn.Del(m.dbi, []byte(key), nil)
	})
	if err != nil {
		m.logger.Error("error deleting value", zap.String("key", key), zap.Error(err))
	}
	return err
}

func (m *MdbxDb) OpenDbi() (*MdbxDb, error) {
	err := m.env.Update(func(txn *mdbx.Txn) error {
		var err error
		m.dbi, err = txn.CreateDBI(m.dbName)
		return err
	})
	return m, err
}
