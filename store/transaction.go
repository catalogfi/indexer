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

func (s *Storage) GetPkScripts(hashes []string, indices []uint32) ([]string, error) {
	keys := make([]string, len(hashes))
	for i, hash := range hashes {
		keys[i] = "pk" + hash + string(indices[i])
	}

	vals, err := s.db.GetMulti(keys)
	if err != nil {
		return nil, err
	}
	scriptPubKeys := make([]string, len(vals))
	for i, val := range vals {
		scriptPubKeys[i] = string(val)
	}
	return scriptPubKeys, nil
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

	batchSize := 25
	wg := sync.WaitGroup{}
	for i := 0; i < len(hashes); i += batchSize {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			end := i + batchSize
			if end > len(hashes) {
				end = len(hashes)
			}
			scriptPubKeys, err := s.GetPkScripts(hashes[i:end], indices[i:end])
			if err != nil {
				s.logger.Error("error getting txs to remove utxos from db", zap.Error(err))
				return
			}
			keys := make([]string, 0)
			vals := make([][]byte, 0)
			for j, pk := range scriptPubKeys {
				keys = append(keys, pk+hashes[i+j]+string(indices[i+j]))
				vals = append(vals, nil)
			}
			err = s.db.DeleteMulti(keys, vals)
			if err != nil {
				s.logger.Error("error deleting utxos from db", zap.Error(err))
				return
			}
		}(i)
	}
	wg.Wait()
	return nil
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
		keys = append(keys, "pk"+utxo.FundingTxHash+string(utxo.FundingTxIndex))
		values = append(values, model.MarshalVout(utxo))
		values = append(values, []byte(utxo.PkScript))
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
