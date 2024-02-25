package command

import (
	"encoding/hex"
	"encoding/json"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/catalogfi/indexer/store"
)

type utxos struct {
	store       *store.Storage
	chainParams *chaincfg.Params
}

func (u *utxos) Name() string {
	return "get_utxos"
}

func (u *utxos) Execute(params json.RawMessage) (interface{}, error) {

	var p string
	err := json.Unmarshal(params, &p)
	if err != nil {
		return nil, err
	}
	address, err := btcutil.DecodeAddress(p, u.chainParams)
	if err != nil {
		return nil, err
	}
	script, err := txscript.PayToAddrScript(address)
	if err != nil {
		return nil, err
	}
	return u.store.GetUTXOs(hex.EncodeToString(script))
}

func UTXOs(store *store.Storage, chainParams *chaincfg.Params) Command {
	return &utxos{
		store:       store,
		chainParams: chainParams,
	}
}

// getTx

type getTx struct {
	store *store.Storage
}

func (g *getTx) Name() string {
	return "get_tx"
}

func (g *getTx) Execute(params json.RawMessage) (interface{}, error) {
	var p string
	err := json.Unmarshal(params, &p)
	if err != nil {
		return nil, err
	}
	tx, exists, err := g.store.GetTx(p)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, store.ErrGetTxNotFound
	}
	return tx, nil
}

func GetTx(store *store.Storage) Command {
	return &getTx{
		store: store,
	}
}

// get_txs_of_address

type getTxsOfAddress struct {
	store       *store.Storage
	chainParams *chaincfg.Params
}

func (g *getTxsOfAddress) Name() string {
	return "get_txs_of_address"
}

func (g *getTxsOfAddress) Execute(params json.RawMessage) (interface{}, error) {
	var p string
	err := json.Unmarshal(params, &p)
	if err != nil {
		return nil, err
	}
	//p is an address, but we need to convert it to a script
	address, err := btcutil.DecodeAddress(p, g.chainParams)
	if err != nil {
		return nil, err
	}
	script, err := txscript.PayToAddrScript(address)
	if err != nil {
		return nil, err
	}
	return g.store.GetTxsOfPubScript(hex.EncodeToString(script))
}

func GetTxsOfAddress(store *store.Storage, chainParams *chaincfg.Params) Command {
	return &getTxsOfAddress{
		store:       store,
		chainParams: chainParams,
	}
}
