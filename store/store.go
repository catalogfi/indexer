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
	"github.com/catalogfi/indexer/command"
	"github.com/catalogfi/indexer/model"
	"github.com/catalogfi/indexer/peer"
	"gorm.io/gorm"
)

// TODO: test reorgs
// TODO: test pending transactions

type Storage interface {
	command.Storage
	peer.Storage
}

type storage struct {
	params *chaincfg.Params
	db     *gorm.DB
}

func NewStorage(params *chaincfg.Params, db *gorm.DB) Storage {
	return &storage{
		params: params,
		db:     db,
	}
}

func (s *storage) Params() *chaincfg.Params {
	return s.params
}

func (s *storage) putTx(tx *wire.MsgTx, block *model.Block, blockIndex uint32) error {
	transactionHash := tx.TxHash().String()
	transaction := &model.Transaction{
		Hash:     transactionHash,
		LockTime: tx.LockTime,
		Version:  tx.Version,
	}

	fOCResult := s.db.FirstOrCreate(transaction, "hash = ?", transactionHash)
	if fOCResult.Error != nil {
		return fOCResult.Error
	}

	if block != nil {
		transaction.BlockID = block.ID
		transaction.BlockIndex = blockIndex
		transaction.BlockHash = block.Hash
		transaction.Block = block
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
			if result := s.db.First(&txInOut, "transaction_out_hash = ? AND transaction_out_index = ?", txIn.PreviousOutPoint.Hash.String(), txIn.PreviousOutPoint.Index); result.Error != nil {
				return result.Error
			}
			txInOut.SpendingTxID = transaction.ID
			txInOut.SpendingTxHash = transactionHash
			txInOut.SpendingTxIndex = inIndex
			txInOut.Sequence = txIn.Sequence
			txInOut.SignatureScript = hex.EncodeToString(txIn.SignatureScript)
			txInOut.Witness = witnessString
			txInOut.SpendingTx = transaction
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

			SpendingTx: transaction,
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

			FundingTx: transaction,
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
		if resp.Error == nil {
			return s.orphanBlock(blockAtHeight)
		}
		return resp.Error
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
		tx.Block = nil
		if resp := s.db.Save(&tx); resp.Error != nil {
			return resp.Error
		}
	}
	resp := s.db.Save(block)
	return resp.Error
}

func (s *storage) BlockExists(blockhash string) bool {
	block := model.Block{}
	res := s.db.First(&block, "hash = ?", blockhash)
	return res.Error != gorm.ErrRecordNotFound
}

func (s *storage) GetPreviousBlockHeight(blockhash string) (int32, error) {
	block := model.Block{}
	if res := s.db.First(&block, "hash = ?", blockhash); res.Error != nil {
		return 0, res.Error
	}

	if res := s.db.First(&block, "hash = ?", block.PreviousBlock); res.Error != nil {
		return 0, res.Error
	}
	return block.Height, nil
}

func (s *storage) PutTx(tx *wire.MsgTx) error {
	return s.putTx(tx, nil, 0)
}

func (s *storage) GetLatestBlockHeight() (int32, error) {
	block := &model.Block{}
	if resp := s.db.Order("height desc").First(block); resp.Error != nil {
		if resp.Error == gorm.ErrRecordNotFound {
			return -1, nil
		}
		return -1, resp.Error
	}
	return block.Height, nil
}

func (s *storage) GetBlockHash(height int32) (string, error) {
	block := &model.Block{}
	if resp := s.db.First(block, "height = ?", height); resp.Error != nil {
		return "", resp.Error
	}
	return block.Hash, nil
}

func (s *storage) GetLatestBlockHash() (string, error) {
	block := &model.Block{}
	if resp := s.db.Order("height desc").First(block); resp.Error != nil {
		return "", resp.Error
	}
	return block.Hash, nil
}

func (s *storage) GetBlockCount() (int32, error) {
	return s.GetLatestBlockHeight()
}

func (s *storage) GetBlockFromHash(blockHash string) (*btcutil.Block, error) {
	block := &model.Block{}
	if resp := s.db.First(block, "hash = ?", blockHash); resp.Error != nil {
		return nil, resp.Error
	}

	prevHash, err := chainhash.NewHashFromStr(block.PreviousBlock)
	if err != nil {
		return nil, err
	}

	merkleRootHash, err := chainhash.NewHashFromStr(block.MerkleRoot)
	if err != nil {
		return nil, err
	}

	blockHeader := wire.NewBlockHeader(block.Version, prevHash, merkleRootHash, block.Bits, block.Nonce)
	blockHeader.Timestamp = block.Timestamp

	msgBlock := wire.NewMsgBlock(blockHeader)

	txs := []model.Transaction{}
	if resp := s.db.Order("block_index").Find(&txs, "block_hash = ?", blockHash); resp.Error != nil {
		return nil, resp.Error
	}
	for _, transaction := range txs {
		tx := wire.NewMsgTx(transaction.Version)
		tx.LockTime = transaction.LockTime
		if err := s.addInputsAndOutputs(transaction.Hash, tx); err != nil {
			return nil, err
		}
		if err := msgBlock.AddTransaction(tx); err != nil {
			return nil, err
		}
	}

	b := btcutil.NewBlock(msgBlock)
	b.SetHeight(block.Height)
	return b, nil
}

func (s *storage) GetHeaderFromHash(blockHash string) (command.BlockHeader, error) {
	block := &model.Block{}
	if resp := s.db.First(block, "hash = ?", blockHash); resp.Error != nil {
		return command.BlockHeader{}, resp.Error
	}
	prevHash, err := chainhash.NewHashFromStr(block.PreviousBlock)
	if err != nil {
		return command.BlockHeader{}, err
	}
	merkleRootHash, err := chainhash.NewHashFromStr(block.MerkleRoot)
	if err != nil {
		return command.BlockHeader{}, err
	}
	blockHeader := wire.NewBlockHeader(block.Version, prevHash, merkleRootHash, block.Bits, block.Nonce)
	blockHeader.Timestamp = block.Timestamp

	var result int64
	if err := s.db.Model(&model.Transaction{}).Where("block_hash = ?", block.Hash).Count(&result).Error; err != nil {
		return command.BlockHeader{}, err
	}

	return command.BlockHeader{
		Header: blockHeader,
		Height: block.Height,
		NumTxs: result,
	}, nil
}

func (s *storage) GetHeaderFromHeight(height int32) (command.BlockHeader, error) {
	block := &model.Block{}
	if resp := s.db.First(block, "height = ?", height); resp.Error != nil {
		return command.BlockHeader{}, resp.Error
	}
	prevHash, err := chainhash.NewHashFromStr(block.PreviousBlock)
	if err != nil {
		return command.BlockHeader{}, err
	}
	merkleRootHash, err := chainhash.NewHashFromStr(block.MerkleRoot)
	if err != nil {
		return command.BlockHeader{}, err
	}
	blockHeader := wire.NewBlockHeader(block.Version, prevHash, merkleRootHash, block.Bits, block.Nonce)
	blockHeader.Timestamp = block.Timestamp

	var result int64
	if err := s.db.Model(&model.Transaction{}).Where("block_hash = ?", block.Hash).Count(&result).Error; err != nil {
		return command.BlockHeader{}, err
	}

	return command.BlockHeader{
		Header: blockHeader,
		Height: block.Height,
		NumTxs: result,
	}, nil
}

func (s *storage) addInputsAndOutputs(txHash string, tx *wire.MsgTx) error {
	txIns := []model.OutPoint{}
	txOuts := []model.OutPoint{}
	if res := s.db.Order("spending_tx_index").Find(&txIns, "spending_tx_hash = ?", txHash); res.Error != nil {
		return res.Error
	}
	for _, txIn := range txIns {
		opHash, err := chainhash.NewHashFromStr(txIn.FundingTxHash)
		if err != nil {
			return fmt.Errorf("invalid op hash: %v", err)
		}

		signatureScript, err := hex.DecodeString(txIn.SignatureScript)
		if err != nil {
			return fmt.Errorf("failed to decode sig script: %v", err)
		}

		witness := strings.Split(txIn.Witness, ",")
		witnessBytes := make([][]byte, len(witness))
		for i := range witness {
			witness, err := hex.DecodeString(witness[i])
			if err != nil {
				return err
			}
			witnessBytes[i] = make([]byte, 32)
			copy(witnessBytes[i], witness)
		}

		tx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(opHash, txIn.FundingTxIndex), signatureScript, witnessBytes))
	}

	if res := s.db.Order("funding_tx_index").Find(&txOuts, "funding_tx_hash = ?", txHash); res.Error != nil {
		return res.Error
	}
	for _, txOut := range txOuts {
		pkScript, err := hex.DecodeString(txOut.PkScript)
		if err != nil {
			return fmt.Errorf("failed to decode pkScript: %v", err)
		}

		tx.AddTxOut(wire.NewTxOut(txOut.Value, pkScript))
	}
	return nil
}

func (s *storage) GetTransaction(txHash string) (command.Transaction, error) {
	transaction := model.Transaction{}
	if res := s.db.Joins("Block").First(&transaction, "transactions.hash = ?", txHash); res.Error != nil {
		return command.Transaction{}, res.Error
	}
	tx := wire.NewMsgTx(transaction.Version)
	tx.LockTime = transaction.LockTime
	if err := s.addInputsAndOutputs(txHash, tx); err != nil {
		return command.Transaction{}, err
	}

	if transaction.Block == nil {
		return command.Transaction{
			Tx: tx,
		}, nil
	}
	return command.Transaction{
		Tx:        tx,
		BlockHash: transaction.BlockHash,
		Height:    transaction.Block.Height,
		BlockTime: transaction.Block.Timestamp.Unix(),
	}, nil
}

func (s *storage) GetBlockLocator() (blockchain.BlockLocator, error) {
	height, err := s.GetLatestBlockHeight()
	if err != nil {
		return nil, err
	}
	locatorIDs := calculateLocator([]int{int(height)})
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

func (storage *storage) ListUnspent(startBlock, endBlock int, addresses []string, includeUnsafe bool, options command.ListUnspentQueryOptions) ([]model.OutPoint, error) {
	outpoints := []model.OutPoint{}
	if !includeUnsafe {
		resp := storage.db.Joins("FundingTx.Block", "height >= ? AND height <= ?", startBlock, endBlock).Joins("FundingTx", "safe = ?", true).Limit(int(options.MaximumCount)).Find(&outpoints, "spender IN ? AND value >= ? AND value <= ?", addresses, options.MinimumAmount, options.MaximumAmount)
		return outpoints, resp.Error
	}
	resp := storage.db.Joins("FundingTx.Block", "height >= ? AND height <= ?", startBlock, endBlock).Joins("FundingTx").Limit(int(options.MaximumCount)).Find(&outpoints, "spender IN ? AND value >= ? AND value <= ?", addresses, options.MinimumAmount, options.MaximumAmount)
	return outpoints, resp.Error
}
