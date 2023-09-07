package zcash

import "go.mongodb.org/mongo-driver/bson/primitive"

type Block struct {
	// mongodb ID
	ID primitive.ObjectID `bson:"_id,omitempty"`

	// blockhash
	Hash     string `bson:"hash"`
	Height   int32
	IsOrphan bool

	// prevBlock Hash
	PreviousBlock string
	Version       int32
	Nonce         uint32
	Timestamp     int64
	Bits          uint32
	MerkleRoot    string
	Transactions  []Transaction

	// miner difficulty
	DifficultyTarget uint32

	//The MagicNumber is typically a fixed value that is agreed upon
	//by the network participants or defined by the protocol. It is used to
	// indicate the beginning of a block and helps to ensure that the block is
	// being interpreted correctly by the network nodes.
	MagicNumber uint32
	BlockSize   uint32
	// final Merkle root of the Sapling note commitment tree
	FinalSaplingRoot string
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

	// Pedersen commitment to the value of the output
	CV string
	// Merkle root of the commitment tree that includes the CV values
	CMU string
	// https://zips.z.cash/zip-0212
	EphemeralKey string
	// ciphertext that represents the encrypted output value.
	EncCiphertext string
	// ciphertext that represents the encrypted output address or script
	OutCiphertext string
	// proof
	ZKProof string
}

type SpendDescription struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	// Pedersen commitment to the value of the output
	CV string
	// value that is used to prove that a specific shielded output has been spent.
	Nullifier string
	// used to get ephemeral key
	ReceivingKey string
	// proof
	ZKProof string

	SpendAuthSig string
}

// joinsplit is when a transaction is made via sheilded addresses
type JoinSplit struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	VPubOld      int64
	VPubNew      int64
	Anchor       string
	Nullifiers   []string
	Commitments  []string
	EphemeralKey string

	RandomSeed string
	//Variable-Length Message Authentication Code
	VMacs   []string
	ZKProof string
}
