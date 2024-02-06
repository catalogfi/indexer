package model

import (
	"time"

	"gorm.io/gorm"
)

type Block struct {
	gorm.Model

	Hash     string `gorm:"index"`
	Height   int32  `gorm:"index"`
	IsOrphan bool   `gorm:"index"`

	PreviousBlock string
	Version       int32
	Nonce         uint32
	Timestamp     time.Time
	Bits          uint32
	MerkleRoot    string
}

type Transaction struct {
	gorm.Model

	Hash     string `gorm:"index"`
	LockTime uint32
	Version  int32
	Safe     bool

	BlockID    uint
	BlockHash  string
	BlockIndex uint32
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
	FundingTxHash  string `gorm:"index"`
	FundingTxIndex uint32 `gorm:"index"`
	PkScript       string
	Value          int64
	Spender        string
	Type           string
}

func NewDB(dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
	db, err := gorm.Open(dialector, opts...)
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Block{}, &Transaction{}, &OutPoint{})
	return db, nil
}
