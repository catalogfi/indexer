package store

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/catalogfi/indexer/model"
	"gorm.io/gorm"
)

func (s *storage) GetBlockLocator() (blockchain.BlockLocator, error) {
	height, err := s.GetLatestBlockHeight()
	if err != nil {
		return nil, err
	}
	locatorIDs := calculateLocator(int64(height))
	blocks := []model.Block{}

	if res := s.db.Find(&blocks, "height in ?", locatorIDs); res.Error != nil {
		return nil, err
	}

	hashes := make([]*chainhash.Hash, len(blocks))
	// indices := make([]int32, len(blocks))
	for i := range blocks {
		hash, err := chainhash.NewHashFromStr(blocks[i].Hash)
		if err != nil {
			return hashes, err
		}
		hashes[i] = hash
		// indices[i] = blocks[i].Height
	}

	// Reverse the list
	for i, j := 0, len(hashes)-1; i < j; i, j = i+1, j-1 {
		hashes[i], hashes[j] = hashes[j], hashes[i]
	}

	return hashes, nil
}

func calculateLocator(topHeight int64) []int {
	var indexes []int

	// Modify the step in the iteration.
	step := int64(1)

	// Start at the top of the chain and work backwards.
	for index := topHeight; index > 0; index -= step {
		// Push top 10 indexes first, then back off exponentially.
		if len(indexes) >= 10 {
			step *= 2
		}

		indexes = append(indexes, int(index))
	}

	// Push the genesis block index.
	indexes = append(indexes, 0)
	return indexes
}

// func calculateLocator(loc []int) []int {

// if len(loc) == 0 {
// 	return []int{}
// }

// height := loc[len(loc)-1]
// if height == 0 {
// 	return loc
// }

// step := 0
// if len(loc) < 12 {
// 	step = 1
// } else {
// 	step = int(math.Pow(2, float64(len(loc)-11)))
// }

// if height <= step {
// 	height = 0
// } else {
// 	height -= step
// }

// return calculateLocator(append(loc, height))
// }

func (s *storage) PutTx(tx *wire.MsgTx) error {
	return s.putTx(tx, nil, 0)
}

func (s *storage) putTx(tx *wire.MsgTx, block *model.Block, blockIndex uint32) error {
	transactionHash := tx.TxHash().String()
	transaction := &model.Transaction{
		Hash:     transactionHash,
		LockTime: tx.LockTime,
		Version:  tx.Version,
	}

	fOCResult := s.db.FirstOrCreate(transaction, model.Transaction{Hash: transactionHash})
	if fOCResult.Error != nil {
		return fOCResult.Error
	}

	if block != nil {
		transaction.BlockID = block.ID
		transaction.BlockIndex = blockIndex
		transaction.BlockHash = block.Hash
		if result := s.db.Save(transaction); result.Error != nil {
			return result.Error
		}
	}

	if fOCResult.RowsAffected == 0 {
		// If the transaction already exists, we don't need to do anything else
		return nil
	}

	for i, txIn := range tx.TxIn {
		inIndex := uint32(i)
		witness := make([]string, len(txIn.Witness))
		for i, w := range txIn.Witness {
			witness[i] = hex.EncodeToString(w)
		}
		witnessString := strings.Join(witness, ",")

		txInOut := model.OutPoint{}
		if txIn.PreviousOutPoint.Hash.String() != "0000000000000000000000000000000000000000000000000000000000000000" && txIn.PreviousOutPoint.Index != 4294967295 {
			// Add SpendingTx to the outpoint
			if result := s.db.First(&txInOut, "funding_tx_hash = ? AND funding_tx_index = ?", txIn.PreviousOutPoint.Hash.String(), txIn.PreviousOutPoint.Index); result.Error != nil {
				return result.Error
			}
			txInOut.SpendingTxID = transaction.ID
			txInOut.SpendingTxHash = transactionHash
			txInOut.SpendingTxIndex = inIndex
			txInOut.Sequence = txIn.Sequence
			txInOut.SignatureScript = hex.EncodeToString(txIn.SignatureScript)
			txInOut.Witness = witnessString
			if res := s.db.Save(&txInOut); res.Error != nil {
				return res.Error
			}
			continue
		}

		// Create coinbase transactions
		if res := s.db.Create(&model.OutPoint{
			SpendingTxID:    transaction.ID,
			SpendingTxHash:  transactionHash,
			SpendingTxIndex: inIndex,
			Sequence:        txIn.Sequence,
			SignatureScript: hex.EncodeToString(txIn.SignatureScript),
			Witness:         witnessString,

			FundingTxHash:  txIn.PreviousOutPoint.Hash.String(),
			FundingTxIndex: txIn.PreviousOutPoint.Index,
		}); res.Error != nil {
			return res.Error
		}
	}

	for i, txOut := range tx.TxOut {
		spenderAddress := ""

		pkScript, err := txscript.ParsePkScript(txOut.PkScript)
		if err == nil {
			addr, err := pkScript.Address(s.params)
			if err != nil {
				return err
			}
			spenderAddress = addr.EncodeAddress()
		}

		// Create a new outpoint
		if res := s.db.Create(&model.OutPoint{
			FundingTxID:    transaction.ID,
			FundingTxHash:  transactionHash,
			FundingTxIndex: uint32(i),
			PkScript:       hex.EncodeToString(txOut.PkScript),
			Value:          txOut.Value,
			Spender:        spenderAddress,
			Type:           pkScript.Class().String(),
		}); res.Error != nil {
			return res.Error
		}
	}
	return nil
}

func (s *storage) PutBlock(block *wire.MsgBlock) error {
	height := int32(-1)
	previousBlock := &model.Block{}
	if block.Header.PrevBlock.String() == s.params.GenesisBlock.BlockHash().String() {
		genesisBlock := btcutil.NewBlock(s.params.GenesisBlock)
		genesisBlock.SetHeight(0)
		if result := s.db.Create(&model.Block{
			Hash:   genesisBlock.Hash().String(),
			Height: 0,

			IsOrphan:      false,
			PreviousBlock: genesisBlock.MsgBlock().Header.PrevBlock.String(),
			Version:       genesisBlock.MsgBlock().Header.Version,
			Nonce:         genesisBlock.MsgBlock().Header.Nonce,
			Timestamp:     genesisBlock.MsgBlock().Header.Timestamp,
			Bits:          genesisBlock.MsgBlock().Header.Bits,
			MerkleRoot:    genesisBlock.MsgBlock().Header.MerkleRoot.String(),
		}); result.Error != nil {
			return result.Error
		}

		// This is created for the coinbase transaction
		resp := s.db.Create(&model.Transaction{
			Hash: "0000000000000000000000000000000000000000000000000000000000000000",
		})
		if resp.Error != nil {
			return resp.Error
		}

		height = 1
	} else {
		if resp := s.db.First(&previousBlock, "hash = ?", block.Header.PrevBlock.String()); resp.Error != nil {
			return resp.Error
		}

		if previousBlock.IsOrphan {
			newlyOrphanedBlock := &model.Block{}
			if resp := s.db.First(newlyOrphanedBlock, "height = ? AND is_orphan = ?", previousBlock.Height, false); resp.Error != nil {
				return resp.Error
			}
			if err := s.orphanBlock(newlyOrphanedBlock); err != nil {
				return err
			}

			previousBlock.IsOrphan = false
			if resp := s.db.Save(&previousBlock); resp.Error != nil {
				return resp.Error
			}
		}

		height = previousBlock.Height + 1
	}

	blockAtHeight := &model.Block{}
	resp := s.db.First(&blockAtHeight, "height = ?", height)
	if resp.Error != gorm.ErrRecordNotFound {
		if resp.Error != nil {
			return resp.Error
		}
		if err := s.orphanBlock(blockAtHeight); err != nil {
			return err
		}
	}

	bblock := &model.Block{
		Hash:   block.Header.BlockHash().String(),
		Height: height,

		IsOrphan:      false,
		PreviousBlock: block.Header.PrevBlock.String(),
		Version:       block.Header.Version,
		Nonce:         block.Header.Nonce,
		Timestamp:     block.Header.Timestamp,
		Bits:          block.Header.Bits,
		MerkleRoot:    block.Header.MerkleRoot.String(),
	}
	if result := s.db.Create(bblock); result.Error != nil {
		return result.Error
	}

	for i, tx := range block.Transactions {
		if err := s.putTx(tx, bblock, uint32(i)); err != nil {
			return err
		}
	}

	aBlock := &model.Block{}
	if resp := s.db.First(aBlock, "hash = ?", block.Header.BlockHash().String()); resp.Error != nil {
		return fmt.Errorf("failed to retrieve the stored block: %v", resp.Error)
	}
	fmt.Println("Block", aBlock.Height, "has been added to the database", aBlock)

	return nil
}

func (s *storage) orphanBlock(block *model.Block) error {
	block.IsOrphan = true

	txs := []model.Transaction{}
	if resp := s.db.Order("block_index").Find(&txs, "block_hash = ?", block.Hash); resp.Error != nil {
		return resp.Error
	}

	for _, tx := range txs {
		tx.BlockID = 0
		tx.BlockHash = ""
		tx.BlockIndex = 0
		if resp := s.db.Save(&tx); resp.Error != nil {
			return resp.Error
		}
	}
	resp := s.db.Save(block)
	return resp.Error
}

func (s *storage) Params() *chaincfg.Params {
	return s.params
}
