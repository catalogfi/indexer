package store

import (
	"fmt"
	"strconv"

	"github.com/catalogfi/indexer/model"
)

var (
	latestBlockHeightKey = "latestBlockHeight"
	orphanKey            = "orphan"
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

func (s *Storage) GetBlock(hash string) (*model.Block, bool, error) {
	data, err := s.db.Get(hash)
	if err != nil {
		if err.Error() == ErrKeyNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	block, err := model.UnmarshalBlock(data)
	if err != nil {
		return nil, false, fmt.Errorf("GetBlock: error unmarshalling block: %w", err)
	}
	return block, true, nil

}

func (s *Storage) GetBlockByHeight(height uint64) (*model.Block, bool, error) {
	data, err := s.db.Get(fmt.Sprint(height))
	if err != nil {
		if err.Error() == ErrKeyNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	block, err := model.UnmarshalBlock(data)
	if err != nil {
		return nil, false, fmt.Errorf("GetBlockByHeight: error unmarshalling block: %w", err)
	}
	return block, true, nil
}

func (s *Storage) GetOrphanBlockByHeight(height uint64) (*model.Block, bool, error) {
	key := fmt.Sprintf("%s_%d", orphanKey, height)
	data, err := s.db.Get(key)
	if err != nil {
		if err.Error() == ErrKeyNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	block, err := model.UnmarshalBlock(data)
	if err != nil {
		return nil, false, fmt.Errorf("GetOrphanBlockByHeight: error unmarshalling block: %w", err)
	}
	return block, true, nil
}

func (s *Storage) GetBlocksRange(start, end uint64, areOrphans bool) ([]*model.Block, error) {
	blocks := make([]*model.Block, 0)
	for i := start; i <= end; i++ {
		var block *model.Block
		var err error
		var exists bool
		if areOrphans {
			block, exists, err = s.GetOrphanBlockByHeight(i)
		} else {
			block, exists, err = s.GetBlockByHeight(i)
		}
		if err != nil {
			return nil, err
		}
		if exists {
			blocks = append(blocks, block)
		}
	}
	return blocks, nil
}

func (s *Storage) BlockExists(hash string) (bool, error) {
	_, err := s.db.Get(hash)
	if (err != nil && err.Error() == ErrKeyNotFound) || err != nil {
		return false, nil
	}
	return true, err
}

func (s *Storage) GetOrphanBlock(hash string) (block *model.Block, exists bool, err error) {
	key := fmt.Sprintf("%s_%s", orphanKey, hash)
	data, err := s.db.Get(key)
	if err != nil {
		if err.Error() == ErrKeyNotFound {
			return nil, false, nil
		}
	}
	block, err = model.UnmarshalBlock(data)
	if err != nil {
		return nil, false, err
	}
	return block, true, nil
}

func (s *Storage) GetBlockTxs(blockHash string, isOrphan bool) ([]*model.Transaction, error) {
	key := blockHash
	if isOrphan {
		key = fmt.Sprintf("%s_%s", orphanKey, blockHash)
	}
	data, err := s.db.Get(key)
	if err != nil {
		return nil, err
	}
	block, err := model.UnmarshalBlock(data)
	if err != nil {
		return nil, err
	}
	return s.GetTxs(block.Txs)

}

func (s *Storage) PutOrphanBlock(block *model.Block) error {
	blockInBytes, err := block.Marshal()
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%s_%s", orphanKey, block.Hash)
	if err = s.db.Put(key, blockInBytes); err != nil {
		return err
	}
	key = fmt.Sprintf("%s_%d", orphanKey, block.Height)
	return s.db.Put(key, blockInBytes)
}

func (s *Storage) PutBlock(block *model.Block) error {
	blockInBytes, err := block.Marshal()
	if err != nil {
		return err
	}
	err = s.db.Put(fmt.Sprint(block.Height), blockInBytes)
	if err != nil {
		return err
	}
	return s.db.Put(block.Hash, blockInBytes)
}
