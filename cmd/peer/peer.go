package main

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/catalogfi/indexer/model"
	"github.com/catalogfi/indexer/peer"
	"github.com/catalogfi/indexer/store"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	db, err := model.NewDB(postgres.Open("postgresql://postgres:QaTl5Rnp0tCIumYH@db.jqpfieqthwnqjxrdnfos.supabase.co:5432/postgres"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// var params *chaincfg.Params
	// switch os.Getenv("NETWORK") {
	// case "mainnet":
	// 	params = &chaincfg.MainNetParams
	// case "testnet":
	// 	params = &chaincfg.TestNet3Params
	// case "regtest":
	// 	params = &chaincfg.RegressionNetParams
	// default:
	// 	panic("invalid network")
	// }
	str := store.NewStorage(&chaincfg.TestNet3Params, db)
	p, err := peer.NewPeer("44.203.96.119:18333", str)
	if err != nil {
		panic(err)
	}
	p.Run()
}
