package command

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
	"github.com/catalogfi/indexer/mongodb/model"
)

type Storage interface {
	GetBlockHash(height int32) (string, error)
	GetLatestBlockHash() (string, error)
	GetLatestBlockHeight() (int32, error)
	GetTransaction(hash string) (Transaction, error)
	GetBlockFromHash(hash string) (*btcutil.Block, error)
	GetHeaderFromHeight(height int32) (BlockHeader, error)
	GetHeaderFromHash(hash string) (BlockHeader, error)
	ListUnspent(startBlock, endBlock int, addresses []string, includeUnsafe bool, queryOptions ListUnspentQueryOptions) ([]model.OutPoint, error)
}

type Command interface {
	Name() string
	Query(str Storage, params []interface{}) (interface{}, error)
}

// getbestblockhash
type getBestBlockHash struct {
}

func GetBestBlockHash() Command {
	return &getBestBlockHash{}
}

func (g *getBestBlockHash) Name() string {
	return "getbestblockhash"
}

func (g *getBestBlockHash) Query(str Storage, params []interface{}) (interface{}, error) {
	return str.GetLatestBlockHash()
}

// getblockhash
type getBlockHash struct {
}

func GetBlockHash() Command {
	return &getBlockHash{}
}

func (g *getBlockHash) Name() string {
	return "getblockhash"
}

func (g *getBlockHash) Query(str Storage, params []interface{}) (interface{}, error) {
	height, ok := params[0].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid parameter type: %T, required a number", params[0])
	}
	return str.GetBlockHash(int32(height))
}

// getblockcount
type getBlockCount struct {
}

func GetBlockCount() Command {
	return &getBlockCount{}
}

func (g *getBlockCount) Name() string {
	return "getblockcount"
}

func (g *getBlockCount) Query(str Storage, params []interface{}) (interface{}, error) {
	return str.GetLatestBlockHeight()
}

// getblockheader
type getBlockHeader struct {
}

func GetBlockHeader() Command {
	return &getBlockHeader{}
}

func (g *getBlockHeader) Name() string {
	return "getblockheader"
}

type BlockHeader struct {
	Header *wire.BlockHeader
	Height int32
	NumTxs int64
}

func (g *getBlockHeader) Query(str Storage, params []interface{}) (interface{}, error) {
	if len(params) < 1 && len(params) > 2 {
		return nil, fmt.Errorf("invalid number of parameters: %d, required 1 or 2", len(params))
	}

	blockHash, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid parameter type: %T, required string", params[0])
	}

	header, err := str.GetHeaderFromHash(blockHash)
	if err != nil {
		return nil, err
	}

	verbose := true
	if len(params) == 2 {
		verbose, ok = params[1].(bool)
		if !ok {
			return nil, fmt.Errorf("invalid parameter type: %T, required boolean", params[1])
		}
	}

	if !verbose {
		buf := new(bytes.Buffer)
		if err := header.Header.Serialize(buf); err != nil {
			return nil, err
		}
		return hex.EncodeToString(buf.Bytes()), nil
	}

	tip, err := str.GetLatestBlockHeight()
	if err != nil {
		return nil, err
	}
	medianHeader, err := str.GetHeaderFromHeight(getMedianBlockHeight(header.Height))
	if err != nil {
		return nil, err
	}
	nextBlockHash, _ := str.GetBlockHash(header.Height + 1)

	return EncodeBlockHeader(header.Header, header.NumTxs, header.Height, uint32(tip-header.Height)+1, medianHeader.Header.Timestamp.Unix(), nextBlockHash)
}

func getMedianBlockHeight(height int32) int32 {
	if height == 0 || height == 1 {
		return height
	}
	if height > 11 {
		return height - 6
	}
	if height%2 == 0 {
		return height / 2
	}
	return (height + 1) / 2
}

// getblock
type getBlock struct {
}

func GetBlock() Command {
	return &getBlock{}
}

func (g *getBlock) Name() string {
	return "getblock"
}

func (g *getBlock) Query(str Storage, params []interface{}) (interface{}, error) {
	if len(params) < 1 && len(params) > 2 {
		return nil, fmt.Errorf("invalid number of parameters: %d, required 1 or 2", len(params))
	}

	blockHash, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid parameter type: %T, required string", params[0])
	}

	block, err := str.GetBlockFromHash(blockHash)
	if err != nil {
		return nil, err
	}

	verbose := 1
	if len(params) == 2 {
		verboseF, ok := params[1].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid parameter type: %T, required boolean", params[1])
		}
		verbose = int(verboseF)
	}

	if verbose == 0 {
		bb, err := block.Bytes()
		if err != nil {
			return nil, err
		}
		return hex.EncodeToString(bb), nil
	}

	tip, err := str.GetLatestBlockHeight()
	if err != nil {
		return nil, err
	}
	medianHeader, err := str.GetHeaderFromHeight(getMedianBlockHeight(block.Height()))
	if err != nil {
		return nil, err
	}
	nextBlockHash, _ := str.GetBlockHash(block.Height() + 1)

	return EncodeBlock(block, uint32(tip-block.Height())+1, medianHeader.Header.Timestamp.Unix(), nextBlockHash, verbose)
}

// getrawtransaction
type getRawTransaction struct {
}

func GetRawTransaction() Command {
	return &getRawTransaction{}
}

func (g *getRawTransaction) Name() string {
	return "getrawtransaction"
}

type Transaction struct {
	Tx        *wire.MsgTx
	BlockHash string
	BlockTime int64
	Height    int32
}

func (g *getRawTransaction) Query(str Storage, params []interface{}) (interface{}, error) {
	txHash := ""
	verbose := false

	switch len(params) {
	case 3:
		fallthrough
	case 2:
		param2, ok := params[1].(bool)
		if !ok {
			return nil, fmt.Errorf("invalid parameter type: %T, required boolean", params[1])
		}
		verbose = param2
		fallthrough
	case 1:
		param1, ok := params[0].(string)
		if !ok {
			return nil, fmt.Errorf("invalid parameter type: %T, required string", params[0])
		}
		txHash = param1
	default:
		return nil, fmt.Errorf("invalid number of parameters: %d, required 1, 2 or 3", len(params))
	}

	tx, err := str.GetTransaction(txHash)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err := tx.Tx.Serialize(buf); err != nil {
		return nil, err
	}
	if !verbose {
		return hex.EncodeToString(buf.Bytes()), nil
	}

	tip, err := str.GetLatestBlockHeight()
	if err != nil {
		return nil, err
	}

	return EncodeTransaction(tx.Tx, tx.BlockHash, uint32(tip-tx.Height)+1, tx.BlockTime), nil
}

// listunspent
type listUnspent struct {
}

func ListUnspent() Command {
	return &listUnspent{}
}

func (g *listUnspent) Name() string {
	return "listunspent"
}

func parseAmount(amt interface{}) (int64, bool) {
	switch amt := amt.(type) {
	case float64:
		return int64(amt * 1e8), true
	case string:
		amtF, err := strconv.ParseFloat(amt, 64)
		if err != nil {
			return 0, false
		}
		return int64(amtF * 1e8), true
	default:
		return 0, false
	}
}

func (g *listUnspent) Query(str Storage, params []interface{}) (interface{}, error) {
	// Defaults
	minconf := 1
	maxconf := 9999999
	addresses := []string{}
	includeUnsafe := true
	queryOptions := ListUnspentQueryOptions{
		MinimumAmount:    0,
		MaximumAmount:    math.MaxInt64,
		MaximumCount:     math.MaxUint32,
		MinimumSumAmount: math.MaxInt64,
	}

	switch len(params) {
	case 5:
		qoj, ok := params[4].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid parameter type: %T, required object", params[4])
		}

		maxAmt, ok := parseAmount(qoj["maximumAmount"])
		if ok {
			queryOptions.MaximumAmount = maxAmt
		}

		minAmt, ok := parseAmount(qoj["minimumAmount"])
		if ok {
			queryOptions.MinimumAmount = minAmt
		}

		maxCount, ok := qoj["maximumCount"].(float64)
		if ok {
			queryOptions.MaximumCount = uint32(maxCount)
		}

		minSumAmt, ok := parseAmount(qoj["minimumSumAmount"])
		if ok {
			queryOptions.MinimumSumAmount = minSumAmt
		}
		fallthrough
	case 4:
		// unconfirmed replacement transactions are considered unsafe
		// and are not eligible for spending by fundrawtransaction and sendtoaddress
		param3, ok := params[3].(bool)
		if !ok {
			return nil, fmt.Errorf("invalid parameter type: %T, required boolean", params[3])
		}
		includeUnsafe = param3
		fallthrough
	case 3:
		// The bitcoin addresses to filter
		param2, ok := params[2].([]interface{}) // only this is used right now
		if !ok {
			return nil, fmt.Errorf("invalid parameter type: %T, required array", params[2])
		}
		addresses = make([]string, len(param2))
		for i, addr := range param2 {
			addresses[i], ok = addr.(string)
			if !ok {
				return nil, fmt.Errorf("invalid parameter type: %T, required string", addr)
			}
		}

		// The maximum confirmations to filter
		param1, ok := params[1].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid parameter type: %T, required number", params[1])
		}
		maxconf = int(param1)

		// The minimum confirmations to filter
		param0, ok := params[0].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid parameter type: %T, required number", params[0])
		}
		minconf = int(param0)
	default:
		return nil, fmt.Errorf("invalid number of parameters needed 3-5, got %d", len(params))
	}

	tip, err := str.GetLatestBlockHeight()
	if err != nil {
		return nil, err
	}

	startBlock := int(tip) + 1 - maxconf
	if startBlock < 0 {
		startBlock = 0
	}

	endBlock := int(tip) + 1 - minconf
	if endBlock < 0 {
		endBlock = 0
	}

	outpoints, err := str.ListUnspent(startBlock, endBlock, addresses, includeUnsafe, queryOptions)
	if err != nil {
		return nil, err
	}

	unspents := []Unspent{}
	sumAmount := int64(0)
	for _, op := range outpoints {
		tx, err := str.GetTransaction(op.FundingTxHash)
		if err != nil {
			return nil, err
		}
		unspents = append(unspents, EncodeUnspent(op, uint32(tip-tx.Height)+1))
		sumAmount += op.Value
		if sumAmount >= queryOptions.MinimumSumAmount {
			break
		}
	}
	return unspents, nil
}
