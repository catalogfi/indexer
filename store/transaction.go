package store

import (
	"fmt"

	"github.com/catalogfi/indexer/model"
	"go.uber.org/zap"
)

func (s *Storage) PutTx(tx *model.Transaction) error {
	return s.db.Put(tx.Hash, tx.Marshal())
}

func (s *Storage) GetTxs(hashes []string) ([]*model.Transaction, error) {
	txs := make([]*model.Transaction, 0)
	for _, hash := range hashes {
		tx, err := s.GetTx(hash)
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

func (s *Storage) RemoveUTXOs(hashs []string, indices []uint32) error {

	if len(hashs) != len(indices) {
		return fmt.Errorf("hashes and indices must have the same length")
	}
	if len(hashs) == 0 {
		s.logger.Info("no utxos to remove")
		return nil
	}

	//get the tx from the db
	txs, err := s.GetTxs(hashs)
	if err != nil {
		return err
	}
	keys := make([]string, 0)
	for i, tx := range txs {
		pkScript := tx.Vouts[indices[i]].PkScript
		keys = append(keys, "IN"+pkScript+string(indices[i]))
	}
	return s.db.DeleteMulti(keys)
}

func (s *Storage) RemoveUTXO(hash string, index uint32) error {
	//get the tx from the db
	tx, err := s.GetTx(hash)
	if err != nil {
		return err
	}
	pkScript := tx.Vouts[index].PkScript

	key := "IN" + pkScript + string(index)
	return s.db.Delete(key)
}

func (s *Storage) GetUTXOs(pkScript string) ([]*model.Vout, error) {
	if len(pkScript) < 10 {
		// if the pkScript is too short, it's not a valid pkScript
		return []*model.Vout{}, nil
	}
	data, err := s.db.Get(pkScript)
	if err != nil {
		return nil, err
	}
	vouts, err := model.UnmarshalVouts(data)
	if err != nil {
		s.logger.Error("error unmarshalling vouts", zap.Error(err), zap.String("pkScript", pkScript), zap.String("data", string(data)))
		return nil, err
	}
	return vouts, nil
}

func (s *Storage) PutUTXOs(utxos []model.Vout) error {
	keys := make([]string, 0)
	values := make([][]byte, 0)
	for _, utxo := range utxos {
		keys = append(keys, "IN"+utxo.PkScript+string(utxo.FundingTxIndex))
		values = append(values, model.MarshalVouts([]*model.Vout{&utxo}))
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

// appends the utxo to the utxos of the pkscript
func (s *Storage) PutUTXO(utxo *model.Vout) error {
	existingUTXOs, err := s.GetUTXOs(utxo.PkScript)
	if err != nil && err.Error() != ErrKeyNotFound {
		return err
	}
	existingUTXOs = append(existingUTXOs, utxo)

	return s.db.Put(utxo.PkScript, model.MarshalVouts(existingUTXOs))
}
