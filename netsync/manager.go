package netsync

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"

	"github.com/catalogfi/indexer/model"
	"github.com/catalogfi/indexer/store"
	"go.uber.org/zap"
)

type SyncManager struct {
	peer         *Peer //TODO: will we have multiple peers in future?
	logger       *zap.Logger
	store        *store.Storage
	chainParams  *chaincfg.Params
	latestHeight uint64
}

type SyncConfig struct {
	PeerAddr    string
	ChainParams *chaincfg.Params
	Store       *store.Storage
	Logger      *zap.Logger
}

func NewSyncManager(config SyncConfig) (*SyncManager, error) {

	logger := config.Logger.Named("syncManager")
	peer, err := NewPeer(config.PeerAddr, config.ChainParams, logger)
	if err != nil {
		return nil, err
	}

	latestHeight, err := config.Store.GetLatestBlockHeight()
	if err != nil && err.Error() != store.ErrKeyNotFound {
		return nil, err
	}

	return &SyncManager{
		peer:         peer,
		chainParams:  config.ChainParams,
		logger:       logger,
		store:        config.Store,
		latestHeight: latestHeight,
	}, nil
}

func (s *SyncManager) Sync() error {

	if err := s.checkForGensisBlock(); err != nil {
		return err
	}

	for {
		ctx, cancel := context.WithCancel(context.Background())
		pendingOnBlocksReq := make(chan struct{})
		go func() {
			s.peer.OnBlock(ctx, func(block *wire.MsgBlock) error {
				if err := s.putBlock(block); err != nil {
					//TODO: handle orphan blocks
					s.logger.Error("error putting block", zap.String("hash", block.BlockHash().String()), zap.Error(err))
					s.logger.Panic("error putting block")
				}
				return nil
			})

			pendingOnBlocksReq <- struct{}{}
		}()

		go s.fetchBlocks()

		s.peer.WaitForDisconnect()
		cancel()
		<-pendingOnBlocksReq
		s.logger.Warn("peer got disconnected... reconnecting")
		reconnectedPeer, err := s.peer.Reconnect()
		if err != nil {
			//TODO: handle reconnection error
			s.logger.Error("error reconnecting peer", zap.Error(err))
			panic(err)
		}
		s.peer = reconnectedPeer
	}

}

func (s *SyncManager) fetchBlocks() {
	for {
		if !s.peer.Connected() {
			break
		}
		locator, err := s.getBlockLocator()
		if err != nil {
			s.logger.Error("error getting latest locator", zap.Error(err))
			continue
		}
		if err := s.peer.PushGetBlocksMsg(locator, &chainhash.Hash{}); err != nil {
			s.logger.Error("error pushing getblocks message", zap.Error(err))
			continue
		}
		limit := 501
		if len(locator) == 0 {
			limit = 500
		}
		for i := 0; i < limit; i++ {
			<-s.peer.fetchBlocksDone
		}
	}

}

func (s *SyncManager) getBlockLocator() ([]*chainhash.Hash, error) {
	latestBlockHeight, err := s.store.GetLatestBlockHeight()
	if err != nil {
		if err.Error() != store.ErrKeyNotFound {
			return nil, err
		}
	}
	//refer to https://en.bitcoin.it/wiki/Protocol_documentation#getblocks
	locatorIDs := calculateLocator(latestBlockHeight)
	blocks, err := s.store.GetBlocks(locatorIDs)
	if err != nil && err.Error() != store.ErrKeyNotFound {
		return nil, err
	}
	hashes := make([]*chainhash.Hash, len(blocks))
	for i := range blocks {
		hash, err := chainhash.NewHashFromStr(blocks[i].Hash)
		if err != nil {
			return hashes, err
		}
		hashes[i] = hash
	}

	return hashes, nil
}

func (s *SyncManager) checkForGensisBlock() error {
	_, exists, err := s.store.GetBlock(s.chainParams.GenesisBlock.BlockHash().String())
	if err != nil {
		return err
	}
	if !exists {
		err := s.putGensisBlock(s.chainParams.GenesisBlock)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SyncManager) putBlock(block *wire.MsgBlock) error {

	height := uint64(0)
	// we check if w already have the block

	exists, err := s.store.BlockExists(block.BlockHash().String())
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	//handle orphan blocks
	_, exists, err = s.store.GetOrphanBlock(block.BlockHash().String())
	if err != nil {
		return err
	}
	if exists {
		// we already have the orphan block too, so just ignore it
		return nil
	}

	previousBlock, exists, err := s.store.GetBlock(block.Header.PrevBlock.String())
	if err != nil {
		return err
	}

	if !exists {
		orphanBlock, exists, err := s.store.GetOrphanBlock(block.Header.PrevBlock.String())
		if err != nil {
			return err
		}
		if exists {
			if s.latestHeight <= orphanBlock.Height+1 {
				// the block we got might not be orphan anymore
				// reorganize the blocks
				err := s.reorganizeBlocks(orphanBlock)
				if err != nil {
					return err
				}
				//proceed with the current block
			} else {
				// we don't have the previous block in the main chain or orphan chain
				// do not process the block
				return s.putOrphanBlock(block, orphanBlock.Height+1)
			}
		} else {
			// we don't have the previous block in the main chain or orphan chain
			// do not process the block
			return nil
		}
	}

	if s.latestHeight >= previousBlock.Height+1 {
		return s.putOrphanBlock(block, previousBlock.Height+1)
	}

	height = previousBlock.Height + 1
	s.logger.Info("processing block", zap.Uint64("height", height), zap.String("hash", block.BlockHash().String()))

	txHashes := make([]string, len(block.Transactions))
	for i, tx := range block.Transactions {
		txHashes[i] = tx.TxHash().String()
	}
	newBlock := model.Block{
		Hash:   block.Header.BlockHash().String(),
		Height: height,

		IsOrphan:      false,
		PreviousBlock: block.Header.PrevBlock.String(),
		Version:       block.Header.Version,
		Nonce:         block.Header.Nonce,
		Timestamp:     block.Header.Timestamp,
		Bits:          block.Header.Bits,
		MerkleRoot:    block.Header.MerkleRoot.String(),
		Txs:           txHashes,
	}
	if err := s.store.PutBlock(&newBlock); err != nil {
		s.logger.Error("error putting block with hash", zap.Error(err))
		return err
	}

	vouts, _, txIns, transactions, err := s.splitTxs(block.Transactions, height, block.BlockHash().String())
	if err != nil {
		return err
	}

	err = s.store.PutUTXOs(vouts)
	if err != nil {
		s.logger.Error("error putting utxos", zap.Error(err))
		return err
	}

	timeNow := time.Now()
	s.logger.Info("putting raw txs")
	err = s.store.PutTxs(transactions)
	if err != nil {
		s.logger.Error("error putting transactions", zap.Error(err))
		return err
	}
	s.logger.Info("putting raw txs done", zap.Duration("time", time.Since(timeNow)))

	timeNow = time.Now()
	hashes := make([]string, 0)
	indices := make([]uint32, 0)
	s.logger.Info("removing utxos")
	for _, in := range txIns {
		if in.PreviousOutPoint.Hash.String() == "0000000000000000000000000000000000000000000000000000000000000000" {
			continue
		}
		hashes = append(hashes, in.PreviousOutPoint.Hash.String())
		indices = append(indices, in.PreviousOutPoint.Index)
	}
	s.logger.Info("removing utxos step 2", zap.Int("len", len(hashes)))
	err = s.store.RemoveUTXOs(hashes, indices)
	if err != nil {
		s.logger.Error("error removing utxos", zap.Error(err))
		return err
	}

	s.logger.Info("removing utxos done", zap.Duration("time", time.Since(timeNow)))

	if err := s.store.SetLatestBlockHeight(height); err != nil {
		return err
	}
	s.logger.Info("successfully block indexed", zap.Uint64("height", height))
	s.latestHeight = height
	return nil
}

func (s *SyncManager) putOrphanBlock(block *wire.MsgBlock, height uint64) error {
	txHashes := make([]string, len(block.Transactions))
	for i, tx := range block.Transactions {
		txHashes[i] = tx.TxHash().String()
	}
	orphanBlock := model.Block{
		Hash:          block.Header.BlockHash().String(),
		Height:        height,
		IsOrphan:      true,
		PreviousBlock: block.Header.PrevBlock.String(),
		Version:       block.Header.Version,
		Nonce:         block.Header.Nonce,
		Timestamp:     block.Header.Timestamp,
		Bits:          block.Header.Bits,
		MerkleRoot:    block.Header.MerkleRoot.String(),
		Txs:           txHashes,
	}
	if err := s.store.PutOrphanBlock(&orphanBlock); err != nil {
		return err
	}

	_, _, _, transactions, err := s.splitTxs(block.Transactions, height, block.BlockHash().String())
	if err != nil {
		return err
	}
	if err = s.store.PutTxs(transactions); err != nil {
		return err
	}
	//we do not put utxos for orphan blocks
	return nil
}

// goal is to travel back to common ancestor of orphan chain and the main chain
// and then reorganize the blocks such that the orphan chain becomes the main chain
// and the main chain becomes the orphan chain (from the common ancestor)
func (s *SyncManager) reorganizeBlocks(orphanBlock *model.Block) error {
	// get the common ancestor of the orphan chain and the main chain
	commonAncestor, exists, err := s.store.GetBlockByHeight(orphanBlock.Height)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("common ancestor block not found")
	}

	// get all blocks from the main chain from the common ancestor
	mainChainBlocks, err := s.store.GetBlocksRange(commonAncestor.Height+1, s.latestHeight, false)
	if err != nil {
		return err
	}

	// get all blocks from the orphan chain from the orphan block
	orphanChainBlocks, err := s.store.GetBlocksRange(orphanBlock.Height+1, s.latestHeight, true)
	if err != nil {
		return err
	}

	// remove all blocks from the main chain from the common ancestor
	// and put them in the orphan chain
	for _, block := range mainChainBlocks {
		if err := s.orphanBlock(block); err != nil {
			return err
		}
	}

	// remove all blocks from the orphan chain from the orphan block
	// and put them in the main chain
	for _, block := range orphanChainBlocks {
		if err := s.unorphanBlock(block); err != nil {
			return err
		}
	}

	return nil

}

func (s *SyncManager) splitTxs(txs []*wire.MsgTx, height uint64, blockHash string) ([]model.Vout, []model.Vin, []*wire.TxIn, []*model.Transaction, error) {

	var vins = make([]model.Vin, 0)
	//TODO: refactor txIns to be in Vins
	var txIns = make([]*wire.TxIn, 0)
	var vouts = make([]model.Vout, 0)
	var transactions = make([]*model.Transaction, len(txs))

	for ti, tx := range txs {
		transactionHash := tx.TxHash().String()
		txVins := make([]model.Vin, len(tx.TxIn))
		txVouts := make([]model.Vout, len(tx.TxOut))
		for i, txIn := range tx.TxIn {
			inIndex := uint32(i)
			witness := make([]string, len(txIn.Witness))
			for i, w := range txIn.Witness {
				witness[i] = hex.EncodeToString(w)
			}
			witnessString := strings.Join(witness, ",")
			if txIn.PreviousOutPoint.Hash.String() != "0000000000000000000000000000000000000000000000000000000000000000" && txIn.PreviousOutPoint.Index != 4294967295 {
				vin := &model.Vin{}
				vin.Sequence = txIn.Sequence
				vin.SignatureScript = hex.EncodeToString(txIn.SignatureScript)
				vin.Witness = witnessString
				vin.SpendingTxHash = transactionHash
				vin.SpendingTxIndex = inIndex
				txVins[i] = *vin
				txIns = append(txIns, txIn)
				continue
			}
			// Create coinbase transactions
			vin := &model.Vin{
				SpendingTxHash:  transactionHash,
				SpendingTxIndex: inIndex,
				Sequence:        txIn.Sequence,
				SignatureScript: hex.EncodeToString(txIn.SignatureScript),
				Witness:         witnessString,
			}
			txVins[i] = *vin
			txIns = append(txIns, txIn)
		}

		for i, txOut := range tx.TxOut {
			spenderAddress := ""

			pkScript, pkErr := txscript.ParsePkScript(txOut.PkScript)
			if pkErr == nil {
				addr, err := pkScript.Address(s.chainParams)
				if err != nil {
					return nil, nil, nil, nil, err
				}
				spenderAddress = addr.EncodeAddress()
			}
			vout := &model.Vout{
				FundingTxHash:  transactionHash,
				FundingTxIndex: uint32(i),
				PkScript:       hex.EncodeToString(txOut.PkScript),
				Value:          txOut.Value,
				Spender:        spenderAddress,
				Type:           pkScript.Class().String(),

				BlockHash:   blockHash,
				BlockHeight: height,
			}
			txVouts[i] = *vout
		}

		transaction := &model.Transaction{
			Hash:     transactionHash,
			LockTime: tx.LockTime,
			Version:  tx.Version,

			BlockHash:   blockHash,
			BlockHeight: height,

			Vins:  txVins,
			Vouts: txVouts,
		}
		vins = append(vins, txVins...)
		vouts = append(vouts, txVouts...)
		transactions[ti] = transaction
	}

	return vouts, vins, txIns, transactions, nil

}

func (s *SyncManager) unorphanBlock(block *model.Block) error {
	block.IsOrphan = false
	txs, err := s.store.GetBlockTxs(block.Hash, true)
	if err != nil {
		return err
	}
	vouts := make([]model.Vout, 0)
	for _, tx := range txs {
		for _, vout := range tx.Vouts {
			vouts = append(vouts, vout)
		}
	}
	if err := s.store.PutUTXOs(vouts); err != nil {
		return err
	}
	return s.store.PutBlock(block)
}

func (s *SyncManager) orphanBlock(block *model.Block) error {
	block.IsOrphan = true
	txs, err := s.store.GetBlockTxs(block.Hash, false)
	if err != nil {
		return err
	}
	hashes := make([]string, 0)
	indices := make([]uint32, 0)
	for _, tx := range txs {
		for _, vin := range tx.Vouts {
			hashes = append(hashes, vin.FundingTxHash)
			indices = append(indices, vin.FundingTxIndex)
		}
	}

	if err := s.store.RemoveUTXOs(hashes, indices); err != nil {
		return err
	}

	return s.store.PutOrphanBlock(block)
}

func (s *SyncManager) putGensisBlock(block *wire.MsgBlock) error {
	genesisBlock := btcutil.NewBlock(s.chainParams.GenesisBlock)
	genesisBlock.SetHeight(0)

	genBlock := &model.Block{
		Hash:          genesisBlock.Hash().String(),
		Height:        0,
		IsOrphan:      false,
		PreviousBlock: genesisBlock.MsgBlock().Header.PrevBlock.String(),
		Version:       genesisBlock.MsgBlock().Header.Version,
		Nonce:         genesisBlock.MsgBlock().Header.Nonce,
		Timestamp:     genesisBlock.MsgBlock().Header.Timestamp,
		Bits:          genesisBlock.MsgBlock().Header.Bits,
		MerkleRoot:    genesisBlock.MsgBlock().Header.MerkleRoot.String(),
		Txs:           []string{"0000000000000000000000000000000000000000000000000000000000000000"},
	}
	if err := s.store.PutBlock(genBlock); err != nil {
		return err
	}

	tx := &model.Transaction{
		Hash: "0000000000000000000000000000000000000000000000000000000000000000",
	}
	if err := s.store.PutTx(tx); err != nil {
		return err
	}
	if err := s.store.SetLatestBlockHeight(0); err != nil {
		return err
	}
	return nil
}

func calculateLocator(topHeight uint64) []uint64 {
	start := int64(topHeight)
	var indexes []uint64
	// Modify the step in the iteration.
	step := int64(1)
	fmt.Println("topHeight", topHeight)
	// Start at the top of the chain and work backwards.
	for index := start; index > 0; index -= step {
		// Push top 10 indexes first, then back off exponentially.
		if len(indexes) >= 10 {
			step *= 2
		}
		indexes = append(indexes, uint64(index))
	}

	// Push the genesis block index.
	indexes = append(indexes, 0)
	return indexes
}
