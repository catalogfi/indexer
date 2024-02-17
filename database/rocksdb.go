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
	bbto := grocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(grocksdb.NewLRUCache(3 << 30))

	opts := grocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)

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

	batchSize := 50
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
	}

	return values, nil
}

func (r *RocksDB) Delete(key string) error {
	wo := grocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()
	return r.db.Delete(wo, []byte(key))
}

func (r *RocksDB) DeleteMulti(keys []string, vals [][]byte) error {

	batchSize := 100
	//delete 500 keys at a time using go routines
	wo := grocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()
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
	batchSize := 250

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
