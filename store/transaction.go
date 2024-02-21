package store

import (
	"fmt"

	"github.com/catalogfi/indexer/model"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (s *Storage) PutTx(tx *model.Transaction) error {
	return s.db.Put(tx.Hash, tx.Marshal())
}

func (s *Storage) GetPkScripts(hashes []string, indices []uint32) ([]string, error) {
	keys := make([]string, len(hashes))
	for i, hash := range hashes {
		keys[i] = getPkKey(hash, indices[i])
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

func (s *Storage) RemoveUTXOs(hashes []string, indices []uint32, vins []model.Vin) error {

	if len(hashes) != len(indices) {
		return fmt.Errorf("hashes and indices must have the same length")
	}
	if len(hashes) == 0 {
		return nil
	}

	batchSize := 100
	eg := new(errgroup.Group)
	for i := 0; i < len(hashes); i += batchSize {
		i := i
		eg.Go(func() error {
			end := i + batchSize
			if end > len(hashes) {
				end = len(hashes)
			}
			scriptPubKeys, err := s.GetPkScripts(hashes[i:end], indices[i:end])
			if err != nil {
				s.logger.Error("error getting txs to remove utxos from db", zap.Error(err))
				return err
			}
			keys := make([]string, len(scriptPubKeys))
			txKeys := make([]string, len(scriptPubKeys))
			txVals := make([][]byte, len(scriptPubKeys))
			for j, pk := range scriptPubKeys {
				keys[j] = pk + hashes[i+j] + string(indices[i+j])
				txKeys[j] = "tx" + pk + vins[i+j].SpendingTxHash
				txVals[j] = []byte(vins[i+j].SpendingTxHash)
			}
			err = s.db.DeleteMulti(keys)
			if err != nil {
				s.logger.Error("error deleting utxos from db", zap.Error(err))
				return err
			}

			if err := s.db.PutMulti(txKeys, txVals); err != nil {
				s.logger.Error("error putting txs to db", zap.Error(err))
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

func (s *Storage) PutUTXOs(utxos []model.Vout) error {
	size := len(utxos) * 2
	keys := make([]string, size)
	values := make([][]byte, size)
	for _, utxo := range utxos {
		key1 := utxo.PkScript + utxo.FundingTxHash + string(utxo.FundingTxIndex)
		key2 := getPkKey(utxo.FundingTxHash, utxo.FundingTxIndex)
		key3 := "tx" + utxo.PkScript + utxo.FundingTxHash
		value1 := model.MarshalVout(utxo)
		value2 := []byte(utxo.PkScript)
		value3 := []byte(utxo.FundingTxHash)
		keys = append(keys, key1, key2, key3)
		values = append(values, value1, value2, value3)
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

func (s *Storage) GetTxs(hashes []string) ([]*model.Transaction, error) {
	data, err := s.db.GetMulti(hashes)
	if err != nil {
		return nil, err
	}
	txs := make([]*model.Transaction, len(data))
	for i, val := range data {
		tx, err := model.UnmarshalTransaction(val)
		if err != nil {
			return nil, err
		}
		txs[i] = tx
	}
	return txs, nil
}

func (s *Storage) PutTxs(txs []*model.Transaction) error {
	keys := make([]string, len(txs))
	values := make([][]byte, len(txs))
	for i, tx := range txs {
		keys[i] = tx.Hash
		values[i] = tx.Marshal()
	}
	return s.db.PutMulti(keys, values)
}

func (s *Storage) GetTxsOfPubScript(scriptPubKey string) ([]*model.Transaction, error) {
	data, err := s.db.GetWithPrefix("tx" + scriptPubKey)
	if err != nil {
		return nil, err
	}
	txHashes := make([]string, len(data))
	for i, val := range data {
		txHashes[i] = string(val)
	}

	return s.GetTxs(txHashes)
}

func getPkKey(hash string, i uint32) string {
	return "pk" + hash + string(i)
}
