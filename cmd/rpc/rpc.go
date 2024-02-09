package main

// import (
// 	"os"

// 	"github.com/btcsuite/btcd/chaincfg"
// 	"github.com/catalogfi/indexer/model"
// 	"github.com/catalogfi/indexer/rpc"
// 	"github.com/catalogfi/indexer/store"
// 	"github.com/gin-gonic/gin"
// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// )

// func main() {
// 	db, err := model.NewDB(postgres.Open(os.Getenv("PSQL_URL")), &gorm.Config{})
// 	if err != nil {
// 		panic(err)
// 	}

// 	var params *chaincfg.Params
// 	switch os.Getenv("NETWORK") {
// 	case "mainnet":
// 		params = &chaincfg.MainNetParams
// 	case "testnet":
// 		params = &chaincfg.TestNet3Params
// 	case "regtest":
// 		params = &chaincfg.RegressionNetParams
// 	default:
// 		panic("invalid network")
// 	}
// 	str := store.NewStorage(params, db)
// 	rpcserver := rpc.Default(str)

// 	s := gin.Default()
// 	s.POST("/", rpcserver.HandleJSONRPC)
// 	s.Run(":8080")
// }
