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
