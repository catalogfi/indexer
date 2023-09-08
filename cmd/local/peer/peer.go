package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/catalogfi/indexer/mongodb"
	"github.com/catalogfi/indexer/mongodb/zcash"
	"github.com/catalogfi/indexer/peer"
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

func PrintPrettyJSON(v interface{}) {
	b, err := json.MarshalIndent(v, "", "	")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
func main() {
	block, err := zcash.GetBlockFromRPC(13586)
	if err != nil {
		panic(err)
	}
	fmt.Println("block")
	PrintPrettyJSON(block.Vtx)
	// fmt.Println(block.)
	return
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

	// bitcoin
	// mongo_db := client.Database("b")
	// str := mongodb.NewStorage(&chaincfg.TestNet3Params, mongo_db)
	// p, err := peer.NewPeer("176.9.63.80:18333", str)
	fmt.Println("start")

	// dogecoin config
	mongo_db := client.Database("dogecoin_testnet")
	str := mongodb.NewStorage(&chaincfg.TestNet3Params, mongo_db)
	p, err := peer.NewPeer("testnets.chain.so:44556", str)
	// zecParams := zec.GetChainParams("test")
	// btcCompatableConfig := &cfg.TestNet3Params
	// printPrettyJSON(btcCompatableConfig)
	// btcCompatableConfig.Name = "testnet"
	// btcCompatableConfig.DNSSeeds = []cfg.DNSSeed{
	// 	{Host: "static.83.80.109.65.clients.your-server.de:18233", HasFiltering: true},
	// 	{Host: "144.22.183.253:18233", HasFiltering: true},
	// 	{Host: "44.1.133.34.bc.googleusercontent.com:18233", HasFiltering: true},
	// 	{Host: "cpe-58-165-97-89.bpw3-r-961.woo.qld.bigpond.net.au:18233", HasFiltering: true},
	// }
	// btcCompatableConfig.Bech32HRPSegwit = "tm"
	// str := mongodb.NewStorage(&dogecoin.DogeCoinTestNet3Params, mongo_db)
	// p, err := peer.NewPeer("testnets.chain.so:44556", str)
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
