package main

import (
	"flag"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/catalogfi/indexer/model"
	"github.com/catalogfi/indexer/peer"
	"github.com/catalogfi/indexer/store"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	token := flag.String("token", "bitcoin", "token")

	flag.Parse()

	fmt.Println(*token)
	db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	str := store.NewStorage(&chaincfg.RegressionNetParams, db)

	p, err := peer.NewPeer(str, *token)
	if err != nil {
		panic(err)
	}
	p.Run()
}
