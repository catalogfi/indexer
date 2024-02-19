package main

import (
	"os"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/catalogfi/indexer/database"
	"github.com/catalogfi/indexer/netsync"
	"github.com/catalogfi/indexer/rpc"
	"github.com/catalogfi/indexer/store"
	"go.uber.org/zap"
)

func main() {

	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stdout"}
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	dbPath := os.Getenv("DB_PATH")

	db, err := database.NewRocksDB(dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var params *chaincfg.Params
	switch os.Getenv("NETWORK") {
	case "mainnet":
		params = &chaincfg.MainNetParams
	case "testnet":
		params = &chaincfg.TestNet3Params
	case "regtest":
		params = &chaincfg.RegressionNetParams
	default:
		panic("invalid network")
	}
	store := store.NewStorage(db).SetLogger(logger)
	syncManager, err := netsync.NewSyncManager(netsync.SyncConfig{
		PeerAddr:    os.Getenv("PEER_URL"),
		ChainParams: params,
		Store:       store,
		Logger:      logger,
	})
	if err != nil {
		panic(err)
	}
	go syncManager.Sync()

	rpcServer := rpc.Default(store, params).SetLogger(logger)
	rpcServer.Run(":8080")
}
