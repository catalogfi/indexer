package main

import (
	"context"
	"encoding/json"
	"fmt"

	cfg "github.com/btcsuite/btcd/chaincfg"
	"github.com/catalogfi/indexer/mongodb"
	"github.com/catalogfi/indexer/peer"
	"github.com/martinboehm/btcutil/chaincfg"
	"github.com/trezor/blockbook/bchain/coins/zec"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// MainNetParams are parser parameters for mainnet
	MainNetParams chaincfg.Params
	// TestNetParams are parser parameters for testnet
	TestNetParams chaincfg.Params
	// RegtestParams are parser parameters for regtest
	RegtestParams chaincfg.Params
)

func GetChainParams(chain string) *chaincfg.Params {
	if !chaincfg.IsRegistered(&MainNetParams) {
		err := chaincfg.Register(&MainNetParams)
		if err == nil {
			err = chaincfg.Register(&TestNetParams)
		}
		if err == nil {
			err = chaincfg.Register(&RegtestParams)
		}
		if err != nil {
			panic(err)
		}
	}
	switch chain {
	case "test":
		return &TestNetParams
	case "regtest":
		return &RegtestParams
	default:
		return &MainNetParams
	}
}

func main() {
	// block, err := zcash.GetBlockFromRPC(0)
	// if err != nil {
	// 	panic(err)
	// }
	// panic(block)

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(context.TODO())

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}

	// mongo_db := client.Database("b")
	// str := mongodb.NewStorage(&chaincfg.TestNet3Params, mongo_db)
	// p, err := peer.NewPeer("176.9.63.80:18333", str)
	fmt.Println("start")

	// dogecoin config
	// mongo_db := client.Database("dogecoin_testnet")
	// str := mongodb.NewStorage(&dogecoin.DogeCoinTestNet3Params, mongo_db)
	// p, err := peer.NewPeer("testnets.chain.so:44556", str)
	zecParams := zec.GetChainParams("test")
	fmt.Println("sez", zecParams)
	jsonData, err := json.MarshalIndent(zecParams, "", "	")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonData))
	mongo_db := client.Database("z")
	// return
	btcCompatableConfig := &cfg.TestNet3Params
	str := mongodb.NewStorage(btcCompatableConfig, mongo_db)
	p, err := peer.NewPeer("cpe-58-165-97-89.bpw3-r-961.woo.qld.bigpond.net.au:18233", str)
	if err != nil {
		panic(err)
	}
	p.Run()
}

// PEERS FOR TESTTING bitcoin testnet
// https://pastebin.com/raw/jxkEpgEq

//PEERS FOR ZCASH
// # 44.1.133.34.bc.googleusercontent.com:18233
// # static.83.80.109.65.clients.your-server.de:18233
// # 144.22.183.253:18233
// # cpe-58-165-97-89.bpw3-r-961.woo.qld.bigpond.net.au:18233
