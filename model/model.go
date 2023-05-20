package model

import (
	"time"

	"gorm.io/gorm"
)

type Block struct {
	gorm.Model

	Hash   string
	Height int32

	PreviousBlock string
	Version       int32
	Nonce         uint32
	Timestamp     time.Time
	Bits          uint32
	MerkleRoot    string
}

type Transaction struct {
	gorm.Model

	Hash     string
	LockTime uint32
	Version  int32
	Safe     bool

	BlockID    uint
	BlockHash  string
	BlockIndex uint32
	Block      *Block `gorm:"foreignKey:BlockID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type OutPoint struct {
	gorm.Model

	SpendingTxID    uint
	SpendingTxHash  string
	SpendingTxIndex uint32
	Sequence        uint32
	SignatureScript string
	Witness         string

	FundingTxID    uint
	FundingTxHash  string
	FundingTxIndex uint32
	PkScript       string
	Value          int64
	Spender        string
	Type           string

	SpendingTx *Transaction `gorm:"foreignKey:SpendingTxID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	FundingTx  *Transaction `gorm:"foreignKey:FundingTxID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func NewDB(dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
	db, err := gorm.Open(dialector, opts...)
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Block{}, &Transaction{}, &OutPoint{})
	return db, nil
}
