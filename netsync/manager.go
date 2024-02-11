package netsync

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

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
	peer        *Peer //TODO: will we have multiple peers in future?
	logger      *zap.Logger
	store       *store.Storage
	chainParams *chaincfg.Params
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

	return &SyncManager{
		peer:        peer,
		chainParams: config.ChainParams,
		logger:      logger,
		store:       config.Store,
	}, nil
}

func (s *SyncManager) Sync() {

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
		fmt.Println("locator", locator)
		if err := s.peer.PushGetBlocksMsg(locator, &chainhash.Hash{}); err != nil {
			s.logger.Error("error pushing getblocks message", zap.Error(err))
			continue
		}
		_, ok := <-s.peer.done
		if !ok {
			break
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

func (s *SyncManager) putBlock(block *wire.MsgBlock) error {
	height := uint64(0)
	if block.Header.PrevBlock.String() == s.chainParams.GenesisBlock.BlockHash().String() {
		s.logger.Info("putting genesis block", zap.String("hash", block.BlockHash().String()))
		if err := s.putGensisBlock(block); err != nil {
			return err
		}
		height = 1
	} else {
		previousBlock, err := s.store.GetBlock(block.Header.PrevBlock.String())
		if err != nil {
			if err.Error() == store.ErrKeyNotFound {
				// we don't have the previous block yet.
				return nil
			}
			return err
		}
		if previousBlock.IsOrphan {
			panic("orphan block")
			//TODO: handle orphan blocks
			// newlyOrphanedBlock, err := s.store.GetBlock(previousBlock.Hash)
			// if err != nil {
			// 	return err
			// }
			// if err := s.orphanBlock(newlyOrphanedBlock); err != nil {
			// 	return err
			// }

			// previousBlock.IsOrphan = false
			// if resp := dbTx.Save(&previousBlock); resp.Error != nil {
			// 	return handleError(resp.Error)
			// }
		}
		height = previousBlock.Height + 1
	}

	existingBlock, err := s.store.GetBlock(block.BlockHash().String())
	if err != nil && err.Error() != store.ErrKeyNotFound {
		return err
	}
	if existingBlock != nil {
		//TODO: handle Orphan blocks
		s.logger.Info("block already exists,todo: handle orphan blocks")
		return nil
	}
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
	s.logger.Info("putting block", zap.Uint64("height", newBlock.Height))
	if err := s.store.PutBlock(&newBlock); err != nil {
		return err
	}

	for _, tx := range block.Transactions {
		if err := s.putTx(tx, block, height); err != nil {
			return err
		}
	}
	if err := s.store.SetLatestBlockHeight(height); err != nil {
		return err
	}
	return nil
}

func (s *SyncManager) putTx(tx *wire.MsgTx, block *wire.MsgBlock, height uint64) error {
	logger := s.logger.Named(fmt.Sprint(height))
	logger.Info("putting tx", zap.String("hash", tx.TxHash().String()))

	transactionHash := tx.TxHash().String()
	transaction := &model.Transaction{
		Hash:     transactionHash,
		LockTime: tx.LockTime,
		Version:  tx.Version,
	}

	existingTx, err := s.store.GetTx(transactionHash)
	if err != nil && err.Error() != store.ErrKeyNotFound {
		return err
	}
	if existingTx != nil {
		return nil
	}

	if block != nil {
		transaction.BlockHash = block.BlockHash().String()
		transaction.BlockHeight = height
	}
	vins := make([]model.Vin, len(tx.TxIn))
	for i, txIn := range tx.TxIn {
		inIndex := uint32(i)
		witness := make([]string, len(txIn.Witness))
		for i, w := range txIn.Witness {
			witness[i] = hex.EncodeToString(w)
		}
		witnessString := strings.Join(witness, ",")

		if txIn.PreviousOutPoint.Hash.String() != "0000000000000000000000000000000000000000000000000000000000000000" && txIn.PreviousOutPoint.Index != 4294967295 {
			if err := s.store.RemoveUTXO(txIn.PreviousOutPoint.Hash.String(), txIn.PreviousOutPoint.Index); err != nil {
				return err
			}
			vin := &model.Vin{}
			vin.Sequence = txIn.Sequence
			vin.SignatureScript = hex.EncodeToString(txIn.SignatureScript)
			vin.Witness = witnessString
			vin.SpendingTxHash = transactionHash
			vin.SpendingTxIndex = inIndex
			vins[i] = *vin
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
		vins[i] = *vin
	}

	vouts := make([]model.Vout, len(tx.TxOut))

	for i, txOut := range tx.TxOut {
		spenderAddress := ""

		pkScript, pkErr := txscript.ParsePkScript(txOut.PkScript)
		if pkErr == nil {
			addr, err := pkScript.Address(s.chainParams)
			if err != nil {
				return err
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

			BlockHash:   block.BlockHash().String(),
			BlockHeight: height,
		}
		//TODO: better validation
		if len(hex.EncodeToString(txOut.PkScript)) > 20 && len(hex.EncodeToString(txOut.PkScript)) < 500 {
			if err := s.store.PutUTXO(vout); err != nil {
				return err
			}
		}

		vouts[i] = *vout
	}

	transaction.Vins = vins
	transaction.Vouts = vouts
	return s.store.PutTx(transaction)
}

func (s *SyncManager) orphanBlock(block *model.Block) error {
	block.IsOrphan = true

	txHashes, err := s.store.GetBlockTxs(block.Hash)
	if err != nil {
		return err
	}

	txs, err := s.store.GetTxs(txHashes)
	if err != nil {
		return err
	}

	for _, tx := range txs {
		tx.BlockHash = ""
		tx.BlockHeight = 0
		err := s.store.PutTx(tx)
		if err != nil {
			return err
		}
	}
	return s.store.PutBlock(block)
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
