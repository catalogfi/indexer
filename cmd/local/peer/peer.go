package main

import (
	"context"
	"fmt"

	"github.com/catalogfi/indexer/mongodb"
	"github.com/catalogfi/indexer/mongodb/dogecoin"
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

	// mongo_db := client.Database("b")
	// str := mongodb.NewStorage(&chaincfg.TestNet3Params, mongo_db)
	// p, err := peer.NewPeer("176.9.63.80:18333", str)
	fmt.Println("start")

	// dogecoin config
	// mongo_db := client.Database("dogecoin_testnet")
	// str := mongodb.NewStorage(&dogecoin.DogeCoinTestNet3Params, mongo_db)
	// p, err := peer.NewPeer("testnets.chain.so:44556", str)
	mongo_db := client.Database("z")
	str := mongodb.NewStorage(&dogecoin.DogeCoinTestNet3Params, mongo_db)
	p, err := peer.NewPeer("zctestseie6wxgio.onion:18233", str)
	if err != nil {
		panic(err)
	}
	p.Run()
}

// PEERS FOR TESTTING bitcoin testnet
// https://pastebin.com/raw/jxkEpgEq
