package store

import (
	"fmt"

	"github.com/catalogfi/indexer/model"
)

var (
	oprphanKey = "orphan"
)

func (s *Storage) PutOrphanTx(tx *model.Transaction) error {

	keys := make([]string, len(tx.Vins)+1)
	values := make([][]byte, len(keys)+1)

	txData, err := tx.Marshal()
	if err != nil {
		return err
	}
	keys[0] = oprphanKey + tx.Hash
	values[0] = txData
	//add all txins to the orphan pool
	for _, vin := range tx.Vins {
		keys = append(keys, oprphanKey+"vin"+vin.TxId+string(vin.Index))
		values = append(values, []byte(tx.Hash))
	}

	return s.db.PutMulti(keys, values)
}

func (s *Storage) GetOrphanTx(hash string) (*model.Transaction, bool, error) {
	data, err := s.db.Get(oprphanKey + hash)
	if err != nil {
		if err.Error() == ErrKeyNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	tx, err := model.UnmarshalTransaction(data)
	if err != nil {
		return nil, false, err
	}
	return tx, true, nil
}

func (s *Storage) GetOrphanDescendants(hash string) ([]*model.Transaction, error) {
	data, err := s.db.GetWithPrefix(oprphanKey + "vin" + hash)
	if err != nil {
		return nil, err
	}
	var txs []*model.Transaction
	for _, d := range data {
		txHash := string(d)
		tx, exists, err := s.GetOrphanTx(txHash)
		if err != nil {
			return nil, err
		}
		if exists {
			txs = append(txs, tx)
		} else {
			return nil, fmt.Errorf("orphan tx not found: %s", txHash)
		}

	}
	return txs, nil
}
