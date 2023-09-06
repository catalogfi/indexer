package zcash

import "go.mongodb.org/mongo-driver/bson/primitive"

type Block struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	Hash     string `bson:"hash"`
	Height   int32
	IsOrphan bool

	PreviousBlock    string
	Version          int32
	Nonce            uint32
	Timestamp        int64
	Bits             uint32
	MerkleRoot       string
	Transactions     []Transaction
	DifficultyTarget uint32
	MagicNumber      uint32
	BlockSize        uint32
	FinalSaplingRoot string
	BlockSolutions   string
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

	Inputs  []Input
	Outputs []Output
	Witness []byte

	VShieldedSpend  []SpendDescription
	VShieldedOutput []OutputDescription
	ValueBalance    int64
	VJoinSplit      []JoinSplit
}

type OutPoint struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	Hash  string
	Index uint32
}

type Input struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	PreviousOutPoint OutPoint
	ScriptLength     uint32
	SignatureScript  string
	Sequence         uint32
}

type Output struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	Value          int64
	PkScriptLength uint32
	PkScript       string
}

type OutputDescription struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	CV            string
	CMU           string
	EphemeralKey  string
	EncCiphertext string
	OutCiphertext string
	ZKProof       string
}

type SpendDescription struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	CV           string
	Anchor       string
	Nullifier    string
	RK           string
	ZKProof      string
	SpendAuthSig string
}

type JoinSplit struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	VPubOld      int64
	VPubNew      int64
	Anchor       string
	Nullifiers   []string
	Commitments  []string
	EphemeralKey string
	RandomSeed   string
	VMacs        []string
	ZKProof      string
}
