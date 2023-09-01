package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Block struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	Hash     string `bson:"hash"`
	Height   int32
	IsOrphan bool

	PreviousBlock string
	Version       int32
	Nonce         uint32
	Timestamp     int64
	Bits          uint32
	MerkleRoot    string
}

type Transaction struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	Hash     string `bson:"hash"`
	LockTime uint32
	Version  int32
	Safe     bool

	BlockID    string
	BlockHash  string
	BlockIndex uint32
}

type OutPoint struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	SpendingTxID    string
	SpendingTxHash  string
	SpendingTxIndex uint32
	Sequence        uint32
	SignatureScript string
	Witness         string

	FundingTxID    string
	FundingTxHash  string
	FundingTxIndex uint32
	PkScript       string
	Value          int64
	Spender        string
	Type           string
}

type Transactions []Transaction

func (transactions Transactions) Len() int {
	return len(transactions)
}

func (transactions Transactions) Less(i, j int) bool {
	return transactions[i].BlockIndex < transactions[j].BlockIndex
}

func (transactions Transactions) Swap(i, j int) {
	transactions[i], transactions[j] = transactions[j], transactions[i]
}

type TxIns []OutPoint

func (txIns TxIns) Len() int {
	return len(txIns)
}

func (txIns TxIns) Less(i, j int) bool {
	return txIns[i].SpendingTxIndex < txIns[j].SpendingTxIndex
}

func (txIns TxIns) Swap(i, j int) {
	txIns[i], txIns[j] = txIns[j], txIns[i]
}

type TxOuts []OutPoint

func (txOuts TxOuts) Len() int {
	return len(txOuts)
}

func (txOuts TxOuts) Less(i, j int) bool {
	return txOuts[i].FundingTxIndex < txOuts[j].FundingTxIndex
}

func (txOuts TxOuts) Swap(i, j int) {
	txOuts[i], txOuts[j] = txOuts[j], txOuts[i]
}
