package mempool

import (
	"encoding/hex"
	"strings"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/catalogfi/indexer/model"
	"github.com/catalogfi/indexer/utils"
)

// package for handling mempool transactions

type storage interface {
	GetTx(hash string) (*model.Transaction, bool, error)
	PutTx(tx *model.Transaction) error
	PutTxs(txs []*model.Transaction) error
	PutUTXOs(vouts []model.Vout) error
	RemoveUTXOs(hashes []string, indices []uint32, vins []model.Vin) error
	PutOrphanTx(tx *model.Transaction) error
	GetOrphanTx(hash string) (*model.Transaction, bool, error)
	GetOrphanDescendants(hash string) ([]*model.Transaction, error)
}

type Mempool struct {
	store storage
}

func New(store storage) *Mempool {
	return &Mempool{
		store: store,
	}
}

// not fully tested. should not be used in production
func (m *Mempool) ProcessTx(tx *wire.MsgTx) error {
	// check if tx is already in mempool
	_, exists, err := m.store.GetTx(tx.TxHash().String())
	if err != nil {
		return err
	}
	if exists {
		//we already have the tx in the mempool, reject
		return nil
	}

	// check if txIns are present in blockchain
	exists, err = m.checkForTxIns(tx)
	if err != nil {
		return err
	}
	if exists {
		// get all descendants of the tx if any
		descendants, err := m.getDescendantsFromOrphanPool(tx)
		if err != nil {
			return err
		}

		err = m.putTx(tx)
		if err != nil {
			return nil
		}
		if len(descendants) > 0 {
			return m.putTxMulti(descendants)
		}
		return nil
	}

	// at this point, we have an orphan tx (it does not have parents)
	// add it to the orphan pool
	return m.putInOrphanPool(tx)
}

// removes the used utxos, adds the new utxos and the tx to the db
func (m *Mempool) putTx(tx *wire.MsgTx) error {

	vouts := make([]model.Vout, len(tx.TxOut))
	for i, txOut := range tx.TxOut {
		pkScript, _ := txscript.ParsePkScript(txOut.PkScript)
		vouts[i] = model.Vout{
			TxId:         tx.TxHash().String(),
			Index:        uint32(i),
			Value:        txOut.Value,
			ScriptPubKey: hex.EncodeToString(txOut.PkScript),
			Type:         pkScript.Class().String(),
		}
	}
	vins := make([]model.Vin, len(tx.TxIn))
	hashes := make([]string, len(tx.TxIn))
	indices := make([]uint32, len(tx.TxIn))
	for i, txIn := range tx.TxIn {
		if txIn.PreviousOutPoint.Hash.String() == "0000000000000000000000000000000000000000000000000000000000000000" && txIn.PreviousOutPoint.Index == 4294967295 {
			continue
		}
		inIndex := uint32(i)
		witness := model.EncodeWitnesss(txIn.Witness)
		vin := &model.Vin{
			TxId:            tx.TxHash().String(),
			Index:           inIndex,
			Sequence:        txIn.Sequence,
			SignatureScript: hex.EncodeToString(txIn.SignatureScript),
			Witness:         witness,
		}
		vins[i] = *vin
		hashes[i] = txIn.PreviousOutPoint.Hash.String()
		indices[i] = txIn.PreviousOutPoint.Index
	}
	transaction := &model.Transaction{
		Hash:      tx.TxHash().String(),
		Version:   tx.Version,
		LockTime:  tx.LockTime,
		Vins:      vins,
		Vouts:     vouts,
		BlockHash: "",
	}

	if err := m.store.PutUTXOs(vouts); err != nil {
		return err
	}
	if err := m.store.RemoveUTXOs(hashes, indices, vins); err != nil {
		return err
	}
	return m.store.PutTx(transaction)
}

func (m *Mempool) putTxMulti(txs []*wire.MsgTx) error {
	vouts, vins, txIns, transactions, err := utils.SplitTxs(txs, "")
	if err != nil {
		return err
	}
	if err := m.store.PutUTXOs(vouts); err != nil {
		return err
	}
	hashes := make([]string, len(txIns))
	indices := make([]uint32, len(txIns))
	for i, txIn := range txIns {
		hashes[i] = txIn.PreviousOutPoint.Hash.String()
		indices[i] = txIn.PreviousOutPoint.Index
	}
	if err := m.store.RemoveUTXOs(hashes, indices, vins[1:]); err != nil {
		return err
	}
	return m.store.PutTxs(transactions)
}

// check if txIns have any txOuts of previous transactions in the indexed data
func (m *Mempool) checkForTxIns(tx *wire.MsgTx) (bool, error) {
	for _, txIn := range tx.TxIn {
		if txIn.PreviousOutPoint.Hash.String() == "0000000000000000000000000000000000000000000000000000000000000000" && txIn.PreviousOutPoint.Index == 4294967295 {
			continue
		}
		// check if txIn is present in blockchain
		_, exists, err := m.store.GetTx(txIn.PreviousOutPoint.Hash.String())
		if err != nil {
			return false, err
		}
		if !exists {
			return false, nil
		}
	}
	return true, nil
}

// we only put the tx in the orphan pool if it does not have any parents
// we do not put utxos or remove utxos from the orphan pool
func (m *Mempool) putInOrphanPool(tx *wire.MsgTx) error {
	txHash := tx.TxHash().String()
	vouts := make([]model.Vout, len(tx.TxOut))
	for i, txOut := range tx.TxOut {
		pkScript, _ := txscript.ParsePkScript(txOut.PkScript)
		vouts[i] = model.Vout{
			TxId:         txHash,
			Index:        uint32(i),
			Value:        txOut.Value,
			ScriptPubKey: hex.EncodeToString(txOut.PkScript),
			Type:         pkScript.Class().String(),
		}
	}
	vins := make([]model.Vin, len(tx.TxIn))
	for i, txIn := range tx.TxIn {
		inIndex := uint32(i)
		witness := make([]string, len(txIn.Witness))
		for i, w := range txIn.Witness {
			witness[i] = hex.EncodeToString(w)
		}
		witnessString := strings.Join(witness, ",")

		if txIn.PreviousOutPoint.Hash.String() != "0000000000000000000000000000000000000000000000000000000000000000" && txIn.PreviousOutPoint.Index != 4294967295 {
			vin := &model.Vin{}
			vin.Sequence = txIn.Sequence
			vin.SignatureScript = hex.EncodeToString(txIn.SignatureScript)
			vin.Witness = witnessString
			vin.TxId = txHash
			vin.Index = inIndex
			vins[i] = *vin
			continue
		}
		// Create coinbase transactions
		vin := &model.Vin{
			TxId:            txHash,
			Index:           inIndex,
			Sequence:        txIn.Sequence,
			SignatureScript: hex.EncodeToString(txIn.SignatureScript),
			Witness:         witnessString,
		}
		vins[i] = *vin
	}

	transaction := &model.Transaction{
		Hash:      txHash,
		Version:   tx.Version,
		LockTime:  tx.LockTime,
		Vins:      vins,
		Vouts:     vouts,
		BlockHash: "",
	}

	return m.store.PutOrphanTx(transaction)
}

func (m *Mempool) getDescendantsFromOrphanPool(tx *wire.MsgTx) ([]*wire.MsgTx, error) {
	txHash := tx.TxHash().String()
	descendants, err := m.store.GetOrphanDescendants(txHash)
	if err != nil {
		return nil, err
	}
	if len(descendants) == 0 {
		return []*wire.MsgTx{}, nil
	}
	wireDescendants := make([]*wire.MsgTx, 0)

descendantsLoop:
	for _, desc := range descendants {
		//validate the descendants
		//need to make sure that the descendant does not have another parent
		//which is not in the indexed data

		for _, vin := range desc.Vins {
			_, exists, err := m.store.GetTx(vin.TxId)
			if err != nil {
				return nil, err
			}
			if !exists {
				continue descendantsLoop
			}
		}

		wireTx, err := desc.ToWireTx()
		if err != nil {
			return nil, err
		}
		descendants, err := m.getDescendantsFromOrphanPool(wireTx)
		if err != nil {
			return nil, err
		}
		wireDescendants = append(wireDescendants, wireTx)
		if len(descendants) > 0 {
			wireDescendants = append(wireDescendants, descendants...)
		}
	}
	return wireDescendants, nil
}
