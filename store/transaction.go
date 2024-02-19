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
			scriptPubKeys, err := s.GetPkScripts(hashes[i:end], indices[i:end])
			if err != nil {
				s.logger.Error("error getting txs to remove utxos from db", zap.Error(err))
				return
			}
			keys := make([]string, len(scriptPubKeys))
			for j, pk := range scriptPubKeys {
				keys[j] = pk + hashes[i+j] + string(indices[i+j])
			}
			err = s.db.DeleteMulti(keys)
			if err != nil {
				s.logger.Error("error deleting utxos from db", zap.Error(err))
				return
			}
		}(i)
	}
	wg.Wait()
	return nil
}

func (s *Storage) PutUTXOs(utxos []model.Vout) error {
	size := len(utxos) * 2
	keys := make([]string, size)
	values := make([][]byte, size)
	for _, utxo := range utxos {
		key1 := utxo.PkScript + utxo.FundingTxHash + string(utxo.FundingTxIndex)
		key2 := "pk" + utxo.FundingTxHash + string(utxo.FundingTxIndex)
		value1 := model.MarshalVout(utxo)
		value2 := []byte(utxo.PkScript)

		keys = append(keys, key1, key2)
		values = append(values, value1, value2)
	}
	return s.db.PutMulti(keys, values)
}

func (s *Storage) GetUTXOs(scriptPubKey string) ([]*model.Vout, error) {
	data, err := s.db.GetWithPrefix(scriptPubKey)
	if err != nil {
		return nil, err
	}
	utxos := make([]*model.Vout, len(data))
	for i, val := range data {
		utxo, err := model.UnmarshalVout(val)
		if err != nil {
			return nil, err
		}
		utxos[i] = utxo
	}
	return utxos, nil
}

func (s *Storage) PutTxs(txs []model.Transaction) error {
	keys := make([]string, len(txs))
	values := make([][]byte, len(txs))
	for i, tx := range txs {
		keys[i] = tx.Hash
		values[i] = tx.Marshal()
	}
	return s.db.PutMulti(keys, values)
}
