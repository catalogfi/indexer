package main

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/catalogfi/indexer/model"
	"github.com/catalogfi/indexer/rpc"
	"github.com/catalogfi/indexer/store"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	str := store.NewStorage(&chaincfg.TestNet3Params, db)
	rpcserver := rpc.Default(str)

	s := gin.Default()
	s.POST("/", rpcserver.HandleJSONRPC)
	s.Run(":8080")
}
