package command

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/catalogfi/indexer/model"
)

// getblockheader
type VerboseBlockHeader struct {
	Hash                 string  `json:"hash"`
	Confirmations        uint32  `json:"confirmations"`
	Height               int32   `json:"height"`
	Version              int32   `json:"version"`
	VersionHex           string  `json:"versionHex"`
	MerkleRoot           string  `json:"merkleroot"`
	Time                 int64   `json:"time"`
	MedianTime           int64   `json:"mediantime"`
	Nonce                uint32  `json:"nonce"`
	Bits                 string  `json:"bits"`
	Difficulty           float64 `json:"difficulty"`
	NumberOfTransactions int64   `json:"nTx"`
	PreviousBlockHash    string  `json:"previousblockhash"`
	NextBlockHash        string  `json:"nextblokchash,omitempty"`
}

func calculateDifficulty(bits uint32) float64 {
	bitBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bitBytes, bits)
	difficultyDen := new(big.Int).Lsh(new(big.Int).SetBytes(bitBytes[1:]), 8*uint(bitBytes[0]-3))
	difficultyNum, ok := new(big.Int).SetString("00000000FFFF0000000000000000000000000000000000000000000000000000", 16)
	if !ok {
		panic("failed to parse difficulty denominator")
	}
	ratio, _ := new(big.Float).Quo(new(big.Float).SetInt(difficultyNum), new(big.Float).SetInt(difficultyDen)).Float64()
	return ratio
}

func EncodeBlockHeader(block *wire.BlockHeader, numTxs int64, height int32, confirmations uint32, medianTime int64, nextBlockHash string) (VerboseBlockHeader, error) {
	return VerboseBlockHeader{
		Hash:                 block.BlockHash().String(),
		Confirmations:        confirmations,
		Height:               height,
		Version:              block.Version,
		VersionHex:           strconv.FormatInt(int64(block.Version), 16),
		MerkleRoot:           block.MerkleRoot.String(),
		Time:                 block.Timestamp.Unix(),
		MedianTime:           medianTime,
		Bits:                 strconv.FormatInt(int64(block.Bits), 16),
		Nonce:                block.Nonce,
		Difficulty:           calculateDifficulty(block.Bits),
		NumberOfTransactions: numTxs,
		PreviousBlockHash:    block.PrevBlock.String(),
		NextBlockHash:        nextBlockHash,
	}, nil
}

// getblock
type VerboseBlock struct {
	Hash                 string      `json:"hash"`
	Confirmations        uint32      `json:"confirmations"`
	Height               int32       `json:"height"`
	Version              int32       `json:"version"`
	VersionHex           string      `json:"versionHex"`
	MerkleRoot           string      `json:"merkleroot"`
	Time                 int64       `json:"time"`
	MedianTime           int64       `json:"mediantime"`
	Nonce                uint32      `json:"nonce"`
	Bits                 string      `json:"bits"`
	Difficulty           float64     `json:"difficulty"`
	NumberOfTransactions int         `json:"nTx"`
	PreviousBlockHash    string      `json:"previousblockhash"`
	NextBlockHash        string      `json:"nextblokchash,omitempty"`
	StrippedSize         int         `json:"strippedsize"`
	Size                 int         `json:"size"`
	Weight               int         `json:"weight"`
	Transactions         interface{} `json:"tx"`
}

func EncodeBlock(block *btcutil.Block, confirmations uint32, medianTime int64, nextBlockHash string, verbose int) (VerboseBlock, error) {
	return VerboseBlock{
		Hash:                 block.Hash().String(),
		Confirmations:        confirmations,
		Size:                 block.MsgBlock().SerializeSize(),
		StrippedSize:         block.MsgBlock().SerializeSizeStripped(),
		Weight:               3*block.MsgBlock().SerializeSizeStripped() + block.MsgBlock().SerializeSize(),
		Height:               block.Height(),
		Version:              block.MsgBlock().Header.Version,
		VersionHex:           strconv.FormatInt(int64(block.MsgBlock().Header.Version), 16),
		MerkleRoot:           block.MsgBlock().Header.MerkleRoot.String(),
		Transactions:         getTxs(block, confirmations, verbose),
		Time:                 block.MsgBlock().Header.Timestamp.Unix(),
		MedianTime:           medianTime,
		Bits:                 strconv.FormatInt(int64(block.MsgBlock().Header.Bits), 16),
		Nonce:                block.MsgBlock().Header.Nonce,
		Difficulty:           calculateDifficulty(block.MsgBlock().Header.Bits),
		NumberOfTransactions: len(block.Transactions()),
		PreviousBlockHash:    block.MsgBlock().Header.PrevBlock.String(),
		NextBlockHash:        nextBlockHash,
	}, nil
}

func getTxs(block *btcutil.Block, confirmations uint32, verbose int) interface{} {
	txs := make([]interface{}, len(block.Transactions()))
	for i, tx := range block.Transactions() {
		if verbose == 1 {
			txs[i] = tx.Hash().String()
		} else {
			txs[i] = EncodeTransaction(tx.MsgTx(), block.Hash().String(), confirmations, block.MsgBlock().Header.Timestamp.Unix())
		}
	}
	return txs
}

// getrawtransaction
type VerboseTransaction struct {
	Hex           string       `json:"hex"`
	TxID          string       `json:"txid"`
	Hash          string       `json:"hash"`
	Size          int          `json:"size"`
	VSize         int          `json:"vsize"`
	Weight        int          `json:"weight"`
	Version       int32        `json:"version"`
	LockTime      uint32       `json:"locktime"`
	VINs          interface{}  `json:"vin"`
	VOUTs         []VerboseOut `json:"vout"`
	BlockHash     string       `json:"blockhash"`
	Confirmations uint32       `json:"confirmations"`
	BlockTime     int64        `json:"blocktime"`
	Time          int64        `json:"time"`
}

type VerboseIn struct {
	TxID        string          `json:"txid"`
	Vout        uint32          `json:"vout"`
	ScriptSig   ScriptSignature `json:"scriptSig"`
	Sequence    uint32          `json:"sequence"`
	TxInWitness []string        `json:"txinwitness"`
}

type Coinbase struct {
	Coinbase    string   `json:"coinbase"`
	TxInWitness []string `json:"txinwitness"`
	Sequence    uint32   `json:"sequence"`
}

type VerboseOut struct {
	Value        float64      `json:"value"`
	Index        uint32       `json:"n"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

type ScriptSignature struct {
	ASM string `json:"asm"`
	HEX string `json:"hex"`
}

type ScriptPubKey struct {
	ASM     string `json:"asm"`
	HEX     string `json:"hex"`
	Type    string `json:"type"`
	Address string `json:"address,omitempty"`
}

func EncodeTransaction(tx *wire.MsgTx, blockHash string, confirmations uint32, time int64) VerboseTransaction {
	buf := new(bytes.Buffer)
	if err := tx.Serialize(buf); err != nil {
		panic(err)
	}
	weight := 3*tx.SerializeSizeStripped() + tx.SerializeSize()
	vsize := int(math.Ceil(float64(weight) / float64(4)))
	return VerboseTransaction{
		Hex:           hex.EncodeToString(buf.Bytes()),
		TxID:          tx.TxHash().String(),
		Hash:          tx.WitnessHash().String(),
		Size:          tx.SerializeSize(),
		VSize:         vsize,
		Weight:        weight,
		VINs:          EncodeVINs(tx.TxIn),
		VOUTs:         EncodeVOUTs(tx.TxOut),
		Version:       tx.Version,
		LockTime:      tx.LockTime,
		BlockHash:     blockHash,
		Confirmations: confirmations,
		Time:          time,
		BlockTime:     time,
	}
}

func EncodeVINs(txins []*wire.TxIn) interface{} {
	if len(txins) == 1 {
		vin := txins[0]
		if vin.PreviousOutPoint.Index == 4294967295 {
			witness := make([]string, len(vin.Witness))
			for j, w := range vin.Witness {
				witness[j] = hex.EncodeToString(w)
			}
			return []Coinbase{{
				Coinbase:    hex.EncodeToString(vin.SignatureScript),
				TxInWitness: witness,
				Sequence:    vin.Sequence,
			}}
		}
	}

	vins := make([]VerboseIn, len(txins))
	for i, vin := range txins {
		witness := make([]string, len(vin.Witness))
		for j, w := range vin.Witness {
			witness[j] = hex.EncodeToString(w)
		}

		asm, err := txscript.DisasmString(vin.SignatureScript)
		if err != nil {
			fmt.Println(err)
		}

		vins[i] = VerboseIn{
			TxID: vin.PreviousOutPoint.Hash.String(),
			Vout: vin.PreviousOutPoint.Index,
			ScriptSig: ScriptSignature{
				ASM: asm,
				HEX: hex.EncodeToString(vin.SignatureScript),
			},
			Sequence:    vin.Sequence,
			TxInWitness: witness,
		}
	}
	return vins
}

func EncodeVOUTs(txouts []*wire.TxOut) []VerboseOut {
	vouts := make([]VerboseOut, len(txouts))
	for i, vout := range txouts {
		asm, err := txscript.DisasmString(vout.PkScript)
		if err != nil {
			panic(err)
		}

		vouts[i] = VerboseOut{
			Value: float64(vout.Value) / float64(100000000),
			Index: uint32(i),
			ScriptPubKey: ScriptPubKey{
				ASM:  asm,
				HEX:  hex.EncodeToString(vout.PkScript),
				Type: "nulldata",
			},
		}

		pks, err := txscript.ParsePkScript(vout.PkScript)
		if err == nil {
			addr, err := pks.Address(&chaincfg.RegressionNetParams)
			if err != nil {
				panic(err)
			}
			vouts[i].ScriptPubKey.Address = addr.EncodeAddress()
			vouts[i].ScriptPubKey.Type = pks.Class().String()
		}
	}
	return vouts
}

// listunspent
type ListUnspentQueryOptionsReq struct {
	MinimumAmount    interface{} `json:"minimumAmount"`
	MaximumAmount    interface{} `json:"maximumAmount"`
	MaximumCount     interface{} `json:"maximumCount"`
	MinimumSumAmount interface{} `json:"minimumSumAmount"`
}

type ListUnspentQueryOptions struct {
	MinimumAmount    int64
	MaximumAmount    int64
	MaximumCount     uint32
	MinimumSumAmount int64
}

type Unspent struct {
	TxID          string  `json:"txid"`
	Vout          uint32  `json:"vout"`
	Address       string  `json:"address"`
	Label         string  `json:"label"`
	Amount        float64 `json:"amount"`
	ScriptPubKey  string  `json:"scriptPubKey"`
	WitnessScript string  `json:"witnessScript"`
	Spendable     bool    `json:"spendable"`
	Solvable      bool    `json:"solvable"`
	Reused        bool    `json:"reused"`
	Description   string  `json:"desc"`
	Confirmations uint32  `json:"confirmations"`
	Safe          bool    `json:"safe"`
}

func EncodeUnspent(op model.OutPoint, tip int32) Unspent {
	return Unspent{
		TxID:          op.FundingTxHash,
		Vout:          op.FundingTxIndex,
		Address:       op.Spender,
		Label:         "",
		Amount:        float64(op.Value) / float64(100000000),
		ScriptPubKey:  op.PkScript,
		WitnessScript: op.Witness,
		Confirmations: uint32(tip-op.FundingTx.Block.Height) + 1,
	}
}
