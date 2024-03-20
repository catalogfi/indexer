package command

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	// "errors"
	"fmt"
	// "os"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
)

// BroadcastCommand represents the command for broadcasting a transaction.
type BroadcastCommand struct {
	RPCURL string
	RPCUser string
	RPCPass string
}

// Name returns the name of the command.
func (b *BroadcastCommand) Name() string {
	return "broadcast"
}

// Execute executes the broadcast command.
func (b *BroadcastCommand) Execute(params json.RawMessage) (interface{}, error) {
	var txHex string
	if err := json.Unmarshal(params, &txHex); err != nil {
		return nil, err
	}

	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, err
	}

	var tx wire.MsgTx
	err = tx.Deserialize(bytes.NewReader(txBytes))
	if err != nil {
		return nil, err
	}

	connCfg := &rpcclient.ConnConfig{
		Host:         b.RPCURL,
		User:         b.RPCUser,
		Pass:         b.RPCPass,
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err
	}
	defer client.Shutdown()

	hash, err := client.SendRawTransaction(&tx, false)
	if err != nil {
		return nil, err
	}

	return fmt.Sprintf("Transaction successfully sent. Transaction hash: %s\n", hash), nil
}

// NewBroadcastCommand creates a new BroadcastCommand with the given configuration.
func NewBroadcastCommand(rpcURL, rpcUser, rpcPass string) *BroadcastCommand {
	return &BroadcastCommand{
		RPCURL: rpcURL,
		RPCUser: rpcUser,
		RPCPass: rpcPass,
	}
}
