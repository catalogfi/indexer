package database

import (
	"fmt"

	"github.com/linxGnu/grocksdb"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type RocksDB struct {
	db     *grocksdb.DB
	logger *zap.Logger
}

func NewRocksDB(path string, logger *zap.Logger) (*RocksDB, error) {
	filter := grocksdb.NewBloomFilter(10)
	bbto := grocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetFilterPolicy(filter)
	bbto.SetOptimizeFiltersForMemory(true)
	bbto.SetBlockCache(grocksdb.NewLRUCache(3 << 30))
	opts := grocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)
	opts.SetUseDirectReads(true)

	db, err := grocksdb.OpenDb(opts, path)
	if err != nil {
		return nil, err
	}
	logger = logger.Named("rocksdb")
	return &RocksDB{
		db:     db,
		logger: logger,
	}, nil
}

func (r *RocksDB) Close() {
	r.db.Close()
}

func (r *RocksDB) Put(key string, value []byte) error {
	wo := grocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()
	return r.db.Put(wo, []byte(key), value)
}

func (r *RocksDB) Get(key string) ([]byte, error) {
	ro := grocksdb.NewDefaultReadOptions()
	defer ro.Destroy()

	slice, err := r.db.Get(ro, []byte(key))
	if err != nil {
		r.logger.Error("error getting key", zap.String("key", key), zap.Error(err))
		return nil, err
	}
	defer slice.Free()
	if !slice.Exists() {
		return nil, fmt.Errorf("key not found")
	}
	val := append([]byte(nil), slice.Data()...)

	return val, nil
}

func (r *RocksDB) GetMulti(keys []string) ([][]byte, error) {
	ro := grocksdb.NewDefaultReadOptions()
	defer ro.Destroy()
	ro.SetFillCache(false)
	batchSize := 150
	values := make([][]byte, len(keys))

	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		keysInBytes := make([][]byte, end-i)
		for j, key := range keys[i:end] {
			keysInBytes[j] = []byte(key)
		}

		slices, err := r.db.MultiGet(ro, keysInBytes...)
		if err != nil {
			r.logger.Error("error getting keys", zap.Strings("keys", keys[i:end]), zap.Error(err))
			return nil, err
		}

		for j, slice := range slices {
			data := make([]byte, len(slice.Data()))
			copy(data, slice.Data())
			values[i+j] = data
			slice.Free()
		}

		keysInBytes = nil
		slices = nil

	}

	return values, nil
}

func (r *RocksDB) Delete(key string) error {
	wo := grocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()
	return r.db.Delete(wo, []byte(key))
}

func (r *RocksDB) DeleteMulti(keys []string) error {

	batchSize := 250
	wo := grocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()
	wo.DisableWAL(true)

	eg := new(errgroup.Group)
	for i := 0; i < len(keys); i += batchSize {
		i := i
		eg.Go(func() error {
			batch := grocksdb.NewWriteBatch()
			defer batch.Destroy()
			for j := i; j < i+batchSize && j < len(keys); j++ {
				batch.Delete([]byte(keys[j]))
			}
			if err := r.db.Write(wo, batch); err != nil {
				return err
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func (r *RocksDB) PutMulti(keys []string, values [][]byte) error {
	batchSize := 500

	wo := grocksdb.NewDefaultWriteOptions()
	wo.DisableWAL(true) // disable write-ahead log
	defer wo.Destroy()

	eg := new(errgroup.Group)

	for i := 0; i < len(keys); i += batchSize {
		i := i
		eg.Go(func() error {
			batch := grocksdb.NewWriteBatch()
			defer batch.Destroy()
			for j := i; j < i+batchSize && j < len(keys); j++ {
				batch.Put([]byte(keys[j]), values[j])
			}
			//write the batch
			if err := r.db.Write(wo, batch); err != nil {
				return err
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		r.logger.Error("error writing batch", zap.Error(err))
		return err
	}
	return nil
}

// GetWithPrefix returns all the values with the given key prefix.
func (r *RocksDB) GetWithPrefix(prefix string) ([][]byte, error) {
	ro := grocksdb.NewDefaultReadOptions()
	defer ro.Destroy()

	ro.SetFillCache(false)
	ro.SetPrefixSameAsStart(true)

	iter := r.db.NewIterator(ro)
	defer iter.Close()

	vals := make([][]byte, 0)
	for iter.Seek([]byte(prefix)); iter.Valid(); iter.Next() {
		key := iter.Key()
		if string(key.Data())[:len(prefix)] != prefix {
			break
		}
		vals = append(vals, append([]byte(nil), iter.Value().Data()...))
		key.Free()
		iter.Value().Free()
	}

	return vals, nil

}
