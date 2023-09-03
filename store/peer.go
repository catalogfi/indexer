package store

import (
	"encoding/hex"
	"fmt"
	"math"
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
	//calculate random block heights like n-1,n-1/2,n-1/4 and so on
	locatorIDs := CalculateLocator([]int{int(height)})
	blocks := []model.Block{}

	if res := s.db.Find(&blocks, "height in ?", locatorIDs); res.Error != nil {
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

func CalculateLocator(loc []int) []int {
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

	return CalculateLocator(append(loc, height))
}

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
	BlockHash := s.params.GenesisBlock.BlockHash().String()
	if block.Header.PrevBlock.String() == BlockHash {
		//this is triggered if first block is created in the blockchain
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
		// resp := s.db.First(&previousBlock, "hash = ?", block.Header.PrevBlock.String())
		// fmt.Println("response: ", resp)
		//this fills the previousBlock with the last block in the database. By matching the hash that we get from the block details from onBlock function.
		if resp := s.db.First(&previousBlock, "hash = ?", block.Header.PrevBlock.String()); resp.Error != nil {
			return resp.Error
		}

		//agar previousBlock orphan  hai toh same height ke unorphaned block ko orphan kar do aut previousBlock ko unorphan kar do
		//so as to continue  the prevblock chain
		// if previousBlock.IsOrphan {
		// 	newlyOrphanedBlock := &model.Block{}
		// 	if resp := s.db.First(newlyOrphanedBlock, "height = ? AND is_orphan = ?", previousBlock.Height, false); resp.Error != nil {
		// 		return resp.Error
		// 	}
		// 	if err := s.orphanBlock(newlyOrphanedBlock); err != nil {
		// 		return err
		// 	}

		// 	previousBlock.IsOrphan = false
		// 	if resp := s.db.Save(&previousBlock); resp.Error != nil {
		// 		return resp.Error
		// 	}
		// }
		if previousBlock.IsOrphan {

			//find the currrnt active chains height
			currHeight, err1 := s.GetLatestUnorphanBlockHeight()
			if err1 != nil {
				return err1
			}
			ActiveChainEnd := &model.Block{}
			//is_orphan check here is redundant as already done by the GetLatestUnorphanBlockHeight function
			if resp := s.db.First(&ActiveChainEnd, "height = ? AND is_orphan = ? ", currHeight, false); resp.Error != nil {
				return resp.Error
			}

			tempBlock := &model.Block{}
			tempBlock = previousBlock
			ActiveChainStart := &model.Block{}
			prevBlock := &model.Block{}
			prevBlock, err1 = s.GetNormalBlockFromHash(tempBlock.PreviousBlock)
			if err1 != nil {
				return err1
			}
			//make the whole required chain unorphaned
			for {
				if !prevBlock.IsOrphan {
					if resp := s.db.First(ActiveChainStart, "height = ? AND is_orphan = ?", tempBlock.Height, false); resp.Error != nil {
						fmt.Printf("no such block exists")
						return resp.Error
					}
					tempBlock.IsOrphan = false
					if resp := s.db.Save(&tempBlock); resp.Error != nil {
						return resp.Error
					}
					break
				}
				tempBlock.IsOrphan = false
				if resp := s.db.Save(&tempBlock); resp.Error != nil {
					return resp.Error
				}
				tempBlock = prevBlock
				prevBlock, err1 = s.GetNormalBlockFromHash(prevBlock.PreviousBlock)
				if err1 != nil {
					return err1
				}
				// tempBlock = tempBlock.PreviousBlock
			}

			//make the prev active chain orphaned
			temp := &model.Block{}
			temp = ActiveChainEnd
			for currHeight >= ActiveChainStart.Height {
				temp.IsOrphan = true
				if resp := s.db.Save(&temp); resp.Error != nil {
					return resp.Error
				}
				temp, err1 = s.GetNormalBlockFromHash(temp.PreviousBlock)
				if err1 != nil {
					return err1
				}
				// temp = temp.PreviousBlock
				currHeight--
			}

		}

		height = previousBlock.Height + 1
		//check if the block height already exists in the database
		sameHeightBlock := &model.Block{}

		sameHeightBlockExists := false
		if resp := s.db.First(&sameHeightBlock, "height = ?", height); resp.Error != nil {
			if resp.Error != gorm.ErrRecordNotFound {
				return resp.Error
			}
		} else {
			sameHeightBlockExists = true
		}

		//setting first time a block as orphan (which is the one at the same height as the block we get)
		//dont just make that block orphan but all blocks which came after that block also orphan
		if !previousBlock.IsOrphan && sameHeightBlockExists {
			for i := height; ; i++ {
				sameHeightBlock := &model.Block{}
				if resp := s.db.First(&sameHeightBlock, "height = ?", i); resp.Error != nil {
					if resp.Error != gorm.ErrRecordNotFound {
						return resp.Error
					}
					break
				}
				sameHeightBlock.IsOrphan = true
				if resp := s.db.Save(&sameHeightBlock); resp.Error != nil {
					return resp.Error
				}
			}
		}
	}

	//yaha aane ke baad joh bhi block h woh sirf insert hoga.

	blockAtHeight := &model.Block{}

	//this line prints this error :
	//2023/06/22 16:02:40 /home/trigo/BITS/CATALOG/BitcoinIndexer/indexer2/store/peer.go:243 record not found

	// This part is actually to deal with the orphan blocks. If the previousBlock is not the last block, it is an orphan block.
	resp := s.db.First(&blockAtHeight, "height = ? AND is_orphan = ?", height, false)
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
	//here the block (bblock) saved in the database
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
	//this logs the output to the console
	fmt.Println("Block", aBlock.Height, "has been added to the database", aBlock)

	return nil
}

// func (s *storage) PutBlock(block *wire.MsgBlock, token string) error {
// 	height := int32(-1)
// 	previousBlock := &model.Block{}
// 	//this is triggered when there are no blocks in the blockchain
// 	fmt.Println("block.Header.PrevBlock.String(): ", block.Header.PrevBlock.String())
// 	fmt.Println("s.params.GenesisBlock.BlockHash().String(): ", s.params.GenesisBlock.BlockHash().String())
// 	// DogeCoinGenesisBlockHash := "3d2160a3b5dc4a9d62e7e66a295f70313ac808440ef7400d6c0772171ce973a5"
// 	if block.Header.PrevBlock.String() == s.params.GenesisBlock.BlockHash().String() {
// 		// if block.Header.PrevBlock.String() == DogeCoinGenesisBlockHash {
// 		genesisBlock := btcutil.NewBlock(s.params.GenesisBlock)
// 		genesisBlock.SetHeight(0)
// 		if result := s.db.Create(&model.Block{
// 			Hash:   genesisBlock.Hash().String(),
// 			Height: 0,

// 			IsOrphan:      false,
// 			PreviousBlock: genesisBlock.MsgBlock().Header.PrevBlock.String(),
// 			Version:       genesisBlock.MsgBlock().Header.Version,
// 			Nonce:         genesisBlock.MsgBlock().Header.Nonce,
// 			Timestamp:     genesisBlock.MsgBlock().Header.Timestamp,
// 			Bits:          genesisBlock.MsgBlock().Header.Bits,
// 			MerkleRoot:    genesisBlock.MsgBlock().Header.MerkleRoot.String(),
// 		}); result.Error != nil {
// 			return result.Error
// 		}

// 		// This is created for the coinbase transaction
// 		resp := s.db.Create(&model.Transaction{
// 			Hash: "0000000000000000000000000000000000000000000000000000000000000000",
// 		})
// 		if resp.Error != nil {
// 			return resp.Error
// 		}

// 		height = 1
// 	} else {
// 		if resp := s.db.First(&previousBlock, "hash = ?", block.Header.PrevBlock.String()); resp.Error != nil {
// 			return resp.Error
// 		}

// 		if previousBlock.IsOrphan {
// 			newlyOrphanedBlock := &model.Block{}
// 			if resp := s.db.First(newlyOrphanedBlock, "height = ? AND is_orphan = ?", previousBlock.Height, false); resp.Error != nil {
// 				return resp.Error
// 			}
// 			if err := s.orphanBlock(newlyOrphanedBlock); err != nil {
// 				return err
// 			}

// 			previousBlock.IsOrphan = false
// 			if resp := s.db.Save(&previousBlock); resp.Error != nil {
// 				return resp.Error
// 			}
// 		}

// 		height = previousBlock.Height + 1
// 	}

// 	blockAtHeight := &model.Block{}
// 	resp := s.db.First(&blockAtHeight, "height = ?", height)
// 	if resp.Error != gorm.ErrRecordNotFound {
// 		if resp.Error == nil {
// 			return s.orphanBlock(blockAtHeight)
// 		}
// 		return resp.Error
// 	}

// 	bblock := &model.Block{
// 		Hash:   block.Header.BlockHash().String(),
// 		Height: height,

// 		IsOrphan:      false,
// 		PreviousBlock: block.Header.PrevBlock.String(),
// 		Version:       block.Header.Version,
// 		Nonce:         block.Header.Nonce,
// 		Timestamp:     block.Header.Timestamp,
// 		Bits:          block.Header.Bits,
// 		MerkleRoot:    block.Header.MerkleRoot.String(),
// 	}
// 	if result := s.db.Create(bblock); result.Error != nil {
// 		return result.Error
// 	}

// 	for i, tx := range block.Transactions {
// 		if err := s.putTx(tx, bblock, uint32(i)); err != nil {
// 			return err
// 		}
// 	}

// 	aBlock := &model.Block{}
// 	if resp := s.db.First(aBlock, "hash = ?", block.Header.BlockHash().String()); resp.Error != nil {
// 		return resp.Error
// 	}
// 	fmt.Println("Block", aBlock.Height, "has been added to the database", aBlock)

// 	return nil
// }

func (s *storage) orphanBlock(block *model.Block) error {
	block.IsOrphan = true

	txs := []model.Transaction{}
	if resp := s.db.Order("block_index").Find(&txs, "block_hash = ?", block.Hash); resp.Error != nil {
		return resp.Error
	}

	//sare transactions ko khali kar do
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
