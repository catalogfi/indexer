package main

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/catalogfi/indexer/mongodb"
	"github.com/catalogfi/indexer/peer"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

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

	mongo_db := client.Database("bitcoin_testnet")
	str := mongodb.NewStorage(&chaincfg.TestNet3Params, mongo_db)
	p, err := peer.NewPeer("44.203.96.119:18333", str)
	fmt.Println("start")

	// dogecoin config
	// str := mongodb.NewStorage(&dogecoin.DogeCoinTestNet3Params, mongo_db)
	// p, err := peer.NewPeer("testnets.chain.so:44556", str)
	if err != nil {
		panic(err)
	}
	p.Run()
}

// PEERS FOR TESTTING bitcoin testnet
// https://pastebin.com/raw/jxkEpgEq
