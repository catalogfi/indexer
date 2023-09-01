package main

import (
	"context"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/catalogfi/indexer/mongodb"
	"github.com/catalogfi/indexer/rpc"
	"github.com/gin-gonic/gin"
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

	mongo_db := client.Database("testdb")
	str := mongodb.NewStorage(&chaincfg.TestNet3Params, mongo_db)
	rpcserver := rpc.Default(str)

	s := gin.Default()
	s.POST("/", rpcserver.HandleJSONRPC)
	s.Run(":8080")
}
