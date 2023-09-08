package zcash

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/pkg/errors"
	"github.com/zcash/lightwalletd/parser"
	"github.com/zcash/lightwalletd/walletrpc"
)

type (
	ZcashRpcReplyGetblock1 struct {
		Hash  string
		Tx    []string
		Trees struct {
			Sapling struct {
				Size uint32
			}
			Orchard struct {
				Size uint32
			}
		}
	}
)

type Options struct {
	RPCUser     string `json:"rpcuser"`
	RPCPassword string `json:"rpcpassword"`
	RPCHost     string `json:"rpchost"`
	RPCPort     string `json:"rpcport"`
}

func PrintPrettyJSON(v interface{}) {
	b, err := json.MarshalIndent(v, "", "	")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
func GetBlockFromRPC(height int) (*walletrpc.CompactBlock, error) {
	var err error
	var opts Options
	var rpcClient *rpcclient.Client
	opts = Options{
		RPCUser:     "qwertyuiop",
		RPCPassword: "qwertyuiop",
		RPCHost:     "127.0.0.1",
		RPCPort:     "8232",
	}
	// https://aws-eu-central-1.json-rpc.cryptoapis.io/nodes/shared/bitcoin/mainnet
	connCfg := &rpcclient.ConnConfig{
		Host:         net.JoinHostPort(opts.RPCHost, opts.RPCPort),
		User:         opts.RPCUser,
		Pass:         opts.RPCPassword,
		HTTPPostMode: true, // Zcash only supports HTTP POST mode
		DisableTLS:   true, // Zcash does not provide TLS by default
	}
	rpcClient, err = rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err
	}
	// if opts.RPCUser != "" && opts.RPCPassword != "" && opts.RPCHost != "" && opts.RPCPort != "" {
	// } else {
	// 	panic("rpcClient is nil")
	// }

	// `block.ParseFromSlice` correctly parses blocks containing v5
	// transactions, but incorrectly computes the IDs of the v5 transactions.
	// We temporarily paper over this bug by fetching the correct txids via a
	// verbose getblock RPC call, which returns the txids.
	//
	// Unfortunately, this RPC doesn't return the raw hex for the block,
	// so a second getblock RPC (non-verbose) is needed (below).

	heightJSON, err := json.Marshal(strconv.Itoa(height))
	if err != nil {
		panic(fmt.Errorf("getBlockFromRPC: %v at height %v", err, height))
	}
	params := make([]json.RawMessage, 2)
	params[0] = heightJSON
	// Fetch the block using the verbose option ("1") because it provides
	// both the list of txids, which we're not yet able to compute for
	// Orchard (V5) transactions, and the block hash (block ID), which
	// we need to fetch the raw data format of the same block. Don't fetch
	// by height in case a reorg occurs between the two getblock calls;
	// using block hash ensures that we're fetching the same block.
	params[1] = json.RawMessage("1")
	result, rpcErr := rpcClient.RawRequest("getblock", params)
	if rpcErr != nil {
		// Check to see if we are requesting a height the zcashd doesn't have yet
		if (strings.Split(rpcErr.Error(), ":"))[0] == "-8" {
			return nil, nil
		}
		return nil, errors.Wrap(rpcErr, "error requesting verbose block")
	}
	var block1 ZcashRpcReplyGetblock1
	err = json.Unmarshal(result, &block1)
	if err != nil {
		return nil, err
	}
	blockHash, err := json.Marshal(block1.Hash)
	if err != nil {
		fmt.Errorf("getBlockFromRPC: %v at height %v", err, height)
	}
	params[0] = blockHash
	params[1] = json.RawMessage("0") // non-verbose (raw hex)
	result, rpcErr = rpcClient.RawRequest("getblock", params)

	// For some reason, the error responses are not JSON
	if rpcErr != nil {
		return nil, errors.Wrap(rpcErr, "error requesting block")
	}

	var blockDataHex string
	err = json.Unmarshal(result, &blockDataHex)
	if err != nil {
		return nil, errors.Wrap(err, "error reading JSON response")
	}
	fmt.Println("BLOCK HEIGHT ", height)
	fmt.Println("block data hex ", blockDataHex)
	blockData, err := hex.DecodeString(blockDataHex)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding getblock output")
	}

	block := parser.NewBlock()
	rest, err := block.ParseFromSlice(blockData)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing block")
	}
	fmt.Println("GetBlockFromRPC")
	if len(rest) != 0 {
		return nil, errors.New("received overlong message")
	}
	if block.GetHeight() != height {
		return nil, errors.New("received unexpected height block")
	}
	for i, t := range block.Transactions() {
		txid, err := hex.DecodeString(block1.Tx[i])
		if err != nil {
			return nil, errors.Wrap(err, "error decoding getblock txid")
		}
		// convert from big-endian
		t.SetTxID(parser.Reverse(txid))
	}
	PrintPrettyJSON(block.Transactions())
	r := block.ToCompact()
	r.ChainMetadata.SaplingCommitmentTreeSize = block1.Trees.Sapling.Size
	r.ChainMetadata.OrchardCommitmentTreeSize = block1.Trees.Orchard.Size
	return r, nil
}
