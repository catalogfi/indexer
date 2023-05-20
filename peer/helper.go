package peer

import (
	"fmt"
	"math"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/catalogfi/indexer/model"
	"gorm.io/gorm"
)

type Storage interface {
	GetBlockLocator() (blockchain.BlockLocator, error)
	PutBlock(block *wire.MsgBlock) error
	Params() *chaincfg.Params
}

func getLatestBlockHeight(db *gorm.DB) (int32, error) {
	block := &model.Block{}
	if resp := db.Order("height desc").First(block); resp.Error != nil {
		if resp.Error == gorm.ErrRecordNotFound {
			return -1, nil
		}
		return -1, fmt.Errorf("db error: %v", resp.Error)
	}
	return block.Height, nil
}

func getBlockLocator(db *gorm.DB) (blockchain.BlockLocator, error) {
	height, err := getLatestBlockHeight(db)
	if err != nil {
		return nil, err
	}
	locatorIDs := calculateLocator([]int{int(height)})
	blocks := []model.Block{}

	if res := db.Find(&blocks, "height in ?", locatorIDs); res.Error != nil {
		return nil, err
	}

	hashes := make([]*chainhash.Hash, len(blocks))
	indices := make([]int32, len(blocks))
	for i := range blocks {
		hash, err := chainhash.NewHashFromStr(blocks[i].Hash)
		if err != nil {
			return hashes, err
		}
		hashes[i] = hash
		indices[i] = blocks[i].Height
	}

	// Reverse the list
	for i, j := 0, len(hashes)-1; i < j; i, j = i+1, j-1 {
		hashes[i], hashes[j] = hashes[j], hashes[i]
	}

	return hashes, nil
}

func calculateLocator(loc []int) []int {
	if len(loc) == 0 {
		return []int{}
	}

	height := loc[len(loc)-1]
	if height == 0 {
		return loc
	}

	step := 0
	if len(loc) < 12 {
		step = 1
	} else {
		step = int(math.Pow(2, float64(len(loc)-11)))
	}

	if height <= step {
		height = 0
	} else {
		height -= step
	}

	return calculateLocator(append(loc, height))
}
