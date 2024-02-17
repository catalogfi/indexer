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
	if slice.Data() == nil {
		return nil, fmt.Errorf("key not found")
	}
	return slice.Data(), nil
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

	//batch 500 keys at a time
	batchSize := 500

	wo := grocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()
	for i := 0; i < len(keys); i += batchSize {

		//create a new batch
		batch := grocksdb.NewWriteBatch()
		defer batch.Destroy()

		//fill the batch
		for j := i; j < i+batchSize && j < len(keys); j++ {
			batch.Put([]byte(keys[j]), values[j])
		}

		//write the batch
		if err := r.db.Write(wo, batch); err != nil {
			return err
		}
	}
	return nil
}

// func Tx(db *RocksDB) (*grocksdb.TransactionDB, error) {
// 	transactionOpts := grocksdb.NewDefaultOptions()
// 	transactionDbOpts := grocksdb.NewDefaultTransactionDBOptions()
// 	transactionDB, err := grocksdb.OpenTransactionDb(transactionOpts, transactionDbOpts, "test")
// 	if err != nil {
// 		return nil, err
// 	}
// 	return transactionDB, nil
// }
