package utils

import (
	"encoding/hex"
	"strings"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/catalogfi/indexer/model"
)

func SplitTxs(txs []*wire.MsgTx, blockHash string) ([]model.Vout, []model.Vin, []*wire.TxIn, []*model.Transaction, error) {

	var vins = make([]model.Vin, 0)
	//TODO: refactor txIns to be in Vins
	var txIns = make([]*wire.TxIn, 0)
	var vouts = make([]model.Vout, 0)
	var transactions = make([]*model.Transaction, len(txs))

	for ti, tx := range txs {
		transactionHash := tx.TxHash().String()
		txVins := make([]model.Vin, len(tx.TxIn))
		txVouts := make([]model.Vout, len(tx.TxOut))
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
				vin.TxId = transactionHash
				vin.Index = inIndex
				txVins[i] = *vin
				txIns = append(txIns, txIn)
				continue
			}
			// Create coinbase transactions
			vin := &model.Vin{
				TxId:            transactionHash,
				Index:           inIndex,
				Sequence:        txIn.Sequence,
				SignatureScript: hex.EncodeToString(txIn.SignatureScript),
				Witness:         witnessString,
			}
			txVins[i] = *vin
			txIns = append(txIns, txIn)
		}

		for i, txOut := range tx.TxOut {
			// we ignore the err from ParsePkScript cause
			// some pkScripts are not standard and we don't care about them
			pkScript, _ := txscript.ParsePkScript(txOut.PkScript)

			vout := &model.Vout{
				TxId:         transactionHash,
				Index:        uint32(i),
				ScriptPubKey: hex.EncodeToString(txOut.PkScript),
				Value:        txOut.Value,

				Type: pkScript.Class().String(),
			}
			txVouts[i] = *vout
		}

		transaction := &model.Transaction{
			Hash:     transactionHash,
			LockTime: tx.LockTime,
			Version:  tx.Version,

			BlockHash: blockHash,

			Vins:  txVins,
			Vouts: txVouts,
		}
		vins = append(vins, txVins...)
		vouts = append(vouts, txVouts...)
		transactions[ti] = transaction
	}

	return vouts, vins, txIns, transactions, nil

}
