package database

import (
	"fmt"
	"sync"

	"github.com/linxGnu/grocksdb"
)

type RocksDB struct {
	db *grocksdb.DB
}

func NewRocksDB(path string) (*RocksDB, error) {
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
	return &RocksDB{
		db: db,
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
	batchSize := 100
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
	//delete 500 keys at a time using go routines
	wo := grocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()
	wo.DisableWAL(true)
	wg := sync.WaitGroup{}
	for i := 0; i < len(keys); i += batchSize {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			batch := grocksdb.NewWriteBatch()
			defer batch.Destroy()
			for j := i; j < i+batchSize && j < len(keys); j++ {
				batch.Delete([]byte(keys[j]))
			}
			if err := r.db.Write(wo, batch); err != nil {
				//TODO: handle error
				panic(err)
			}
		}(i)
	}
	wg.Wait()
	return nil

}

func (r *RocksDB) PutMulti(keys []string, values [][]byte) error {
	batchSize := 500

	wo := grocksdb.NewDefaultWriteOptions()
	wo.DisableWAL(true) // disable write-ahead log
	defer wo.Destroy()
	wg := sync.WaitGroup{}

	for i := 0; i < len(keys); i += batchSize {
		wg.Add(1)
		go func(i int) {
			batch := grocksdb.NewWriteBatch()
			for j := i; j < i+batchSize && j < len(keys); j++ {
				batch.Put([]byte(keys[j]), values[j])
			}

			//write the batch
			if err := r.db.Write(wo, batch); err != nil {
				//TODO: handle error
				panic(err)
			}
			wg.Done()
			batch.Destroy()
		}(i)
	}
	wg.Wait()
	return nil
}

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
	}

	return vals, nil

}
