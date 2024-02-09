package store

import (
	"github.com/catalogfi/indexer/model"
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

func (s *Storage) RemoveUTXO(hash string, index uint32) error {
	//get the tx from the db
	tx, err := s.GetTx(hash)
	if err != nil {
		return err
	}
	pkScript := tx.Vouts[index].PkScript

	//get the utxos from the db
	utxos, err := s.GetUTXOs(pkScript)
	if err != nil {
		return err
	}
	//remove the utxo
	for i, utxo := range utxos {
		if utxo.FundingTxHash == hash && utxo.FundingTxIndex == index {
			utxos = append(utxos[:i], utxos[i+1:]...)
			break
		}
	}
	//put the utxos back in the db
	return s.db.Put(pkScript, model.MarshalVouts(utxos))
}

func (s *Storage) GetUTXOs(pkScript string) ([]*model.Vout, error) {
	data, err := s.db.Get(pkScript)
	if err != nil {
		return nil, err
	}
	return model.UnmarshalVouts(data)
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
