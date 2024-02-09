package model

import (
	"encoding/json"
	"time"
)

type Block struct {
	Hash     string
	Height   uint64
	IsOrphan bool

	PreviousBlock string
	Version       int32
	Nonce         uint32
	Timestamp     time.Time
	Bits          uint32
	MerkleRoot    string

	//Transactions
	Txs []string
}

type Transaction struct {
	Hash     string
	LockTime uint32
	Version  int32
	Safe     bool

	BlockHash   string
	BlockHeight uint64

	//VINs
	Vins []Vin
	//VOUTs
	Vouts []Vout
}

type Vin struct {
	SpendingTxHash  string
	SpendingTxIndex uint32
	Sequence        uint32
	SignatureScript string
	Witness         string
}

type Vout struct {
	FundingTxHash  string
	FundingTxIndex uint32
	PkScript       string
	Value          int64
	Spender        string
	Type           string

	//Block info
	BlockHash   string
	BlockHeight uint64
}

func UnmarshalBlock(data []byte) (*Block, error) {
	block := &Block{}
	err := json.Unmarshal(data, block)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (b *Block) Marshal() []byte {
	data, _ := json.Marshal(b)
	return data
}

func UnmarshalVout(data []byte) (*Vout, error) {
	vout := &Vout{}
	err := json.Unmarshal(data, vout)
	if err != nil {
		return nil, err
	}
	return vout, nil
}

func UnmarshalVouts(data []byte) ([]*Vout, error) {
	var vouts []*Vout
	err := json.Unmarshal(data, &vouts)
	if err != nil {
		return nil, err
	}
	return vouts, nil
}

func MarshalVouts(vouts []*Vout) []byte {
	data, _ := json.Marshal(vouts)
	return data
}

func (v *Vout) Marshal() []byte {
	data, _ := json.Marshal(v)
	return data
}

func (t *Transaction) Marshal() []byte {
	data, _ := json.Marshal(t)
	return data
}

func UnmarshalTransaction(data []byte) (*Transaction, error) {
	tx := &Transaction{}
	err := json.Unmarshal(data, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
