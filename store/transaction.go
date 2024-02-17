package store

import (
	"fmt"
	"sync"

	"github.com/catalogfi/indexer/model"
	"go.uber.org/zap"
)

func (s *Storage) PutTx(tx *model.Transaction) error {
	return s.db.Put(tx.Hash, tx.Marshal())
}

func (s *Storage) GetTxs(hashes []string) ([]*model.Transaction, error) {
	data, err := s.db.GetMulti(hashes)
	if err != nil {
		return nil, fmt.Errorf("error getting txs: %w", err)
	}
	txs := make([]*model.Transaction, 0)
	for _, d := range data {
		tx, err := model.UnmarshalTransaction(d)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func (s *Storage) GetTx(hash string) (*model.Transaction, error) {
	data, err := s.db.Get(hash)
	if err != nil {
		return nil, err
	}
	return model.UnmarshalTransaction(data)
}

func (s *Storage) RemoveUTXOs(hashes []string, indices []uint32) error {

	if len(hashes) != len(indices) {
		return fmt.Errorf("hashes and indices must have the same length")
	}
	if len(hashes) == 0 {
		return nil
	}
	s.logger.Info("getting txs to remove utxos from db")

	batchSize := 50
	wg := sync.WaitGroup{}
	for i := 0; i < len(hashes); i += batchSize {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			end := i + batchSize
			if end > len(hashes) {
				end = len(hashes)
			}
			txs, err := s.GetTxs(hashes[i:end])
			if err != nil {
				s.logger.Error("error getting txs to remove utxos from db", zap.Error(err))
				return
			}
			keys := make([]string, 0)
			vals := make([][]byte, 0)
			for j, tx := range txs {
				pkScript := tx.Vouts[indices[i+j]].PkScript
				keys = append(keys, pkScript+hashes[i+j]+string(indices[i+j]))
				vals = append(vals, nil)
			}
			// free the memory
			txs = nil
			err = s.db.DeleteMulti(keys, vals)
			if err != nil {
				s.logger.Error("error deleting utxos from db", zap.Error(err))
				return
			}
		}(i)
	}
	wg.Wait()
	return nil

	//get the tx from the db
	// txs, err := s.GetTxs(hashes)
	// if err != nil {
	// 	return err
	// }
	// s.logger.Info("got txs to remove utxos from db")
	// keys := make([]string, 0)
	// vals := make([][]byte, 0)
	// for i, tx := range txs {
	// 	pkScript := tx.Vouts[indices[i]].PkScript
	// 	keys = append(keys, pkScript+hashes[i]+string(indices[i]))
	// 	vals = append(vals, nil)
	// }
	// // free the memory
	// txs = nil
	// return s.db.DeleteMulti(keys, vals)
}

func (s *Storage) RemoveUTXO(hash string, index uint32) error {
	//get the tx from the db
	tx, err := s.GetTx(hash)
	if err != nil {
		return err
	}
	pkScript := tx.Vouts[index].PkScript

	key := pkScript + hash + string(index)
	return s.db.Delete(key)
}

func (s *Storage) PutUTXOs(utxos []model.Vout) error {
	keys := make([]string, 0)
	values := make([][]byte, 0)
	for _, utxo := range utxos {
		keys = append(keys, utxo.PkScript+utxo.FundingTxHash+string(utxo.FundingTxIndex))
		values = append(values, model.MarshalVout(utxo))
	}
	return s.db.PutMulti(keys, values)
}

func (s *Storage) PutTxs(txs []model.Transaction) error {
	keys := make([]string, 0)
	values := make([][]byte, 0)
	for _, tx := range txs {
		keys = append(keys, tx.Hash)
		values = append(values, tx.Marshal())
	}
	return s.db.PutMulti(keys, values)
}
