package store

import (
	"fmt"
	"strconv"

	"github.com/catalogfi/indexer/model"
)

var (
	latestBlockHeightKey = "latestBlockHeight"
)

// GetLatestBlockHeight returns the latest block height in the database
func (s *Storage) GetLatestBlockHeight() (uint64, error) {
	data, err := s.db.Get(latestBlockHeightKey)
	if err != nil {
		return 0, err
	}
	height, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, err
	}
	return uint64(height), nil
}

func (s *Storage) SetLatestBlockHeight(height uint64) error {
	heightStr := strconv.Itoa(int(height))
	return s.db.Put(latestBlockHeightKey, []byte(heightStr))
}

// GetBlocks returns the blocks with the given heights.
func (s *Storage) GetBlocks(heights []uint64) ([]*model.Block, error) {
	blocks := make([]*model.Block, 0)
	for _, height := range heights {
		data, err := s.db.Get(fmt.Sprint(height))
		if err != nil {
			return nil, err
		}
		block, err := model.UnmarshalBlock(data)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil

}

func (s *Storage) GetBlock(hash string) (*model.Block, error) {
	data, err := s.db.Get(hash)
	if err != nil {
		return nil, err
	}
	return model.UnmarshalBlock(data)
}

func (s *Storage) GetOrphanBlock(height uint64) (*model.Block, error) {
	panic("not implemented")
}

func (s *Storage) GetBlockTxs(blockHash string) ([]string, error) {
	data, err := s.db.Get(blockHash)
	if err != nil {
		return nil, err
	}
	block, err := model.UnmarshalBlock(data)
	if err != nil {
		return nil, err
	}
	return block.Txs, nil
}

func (s *Storage) PutBlock(block *model.Block) error {
	err := s.db.Put(fmt.Sprint(block.Height), block.Marshal())
	if err != nil {
		return err
	}
	return s.db.Put(block.Hash, block.Marshal())
}
