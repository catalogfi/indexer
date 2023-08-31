package mongodb

import (
	"context"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/catalogfi/indexer/command"
	"github.com/catalogfi/indexer/mongodb/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *storage) GetPreviousBlockHeight(blockhash string) (int32, error) {
	block := model.Block{}
	if err := s.db.Collection("blocks").FindOne(context.TODO(), bson.D{{Key: "hash", Value: blockhash}}).Decode(&block); err != nil {
		return 0, err
	}
	return block.Height, nil
}

func (s *storage) GetLatestBlockHeight() (int32, error) {
	block := model.Block{}
	err := s.db.Collection("blocks").FindOne(context.TODO(), bson.M{} , &options.FindOneOptions{
		Sort: bson.D{{Key: "height", Value: -1}},
	}).Decode(&block)
	if err != nil {
		if err == mongo.ErrNoDocuments{
			return -1, nil
		}
		return -1, err
	}
	return block.Height, nil
}

func (s *storage) GetBlockHash(height int32) (string, error) {
	block := model.Block{}
	if err := s.db.Collection("blocks").FindOne(context.TODO(), bson.D{{Key: "height", Value: height}}).Decode(&block); err != nil {
		return "", err
	}
	return block.Hash, nil
}

func (s *storage) GetLatestBlockHash() (string, error) {
	block := &model.Block{}
	// s.db.Collection("blocks").InsertOne()

	err := s.db.Collection("blocks").FindOne(context.TODO(), bson.D{} , &options.FindOneOptions{
		Sort: bson.D{{Key: "height", Value: -1}},
	}).Decode(&block)
	if err != nil {
		if err == mongo.ErrNoDocuments{
			return "", nil
		}
		return "", err
	}
	return block.Hash, nil
}

func (s *storage) GetBlockCount() (int32, error) {
	return s.GetLatestBlockHeight()
}

func (s *storage) GetBlockFromHash(blockhash string) (*btcutil.Block, error) {
	block := model.Block{}
	if err := s.db.Collection("blocks").FindOne(context.TODO(), bson.D{{Key: "hash", Value: blockhash}}).Decode(&block); err != nil {
		return nil, err
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
	blockHeader.Timestamp = time.Unix(block.Timestamp, 0)

	msgBlock := wire.NewMsgBlock(blockHeader)

	txs := model.Transactions{}
	cursor, err := s.db.Collection("transactions").Find(context.TODO(), bson.D{{Key: "block_hash", Value: blockhash}})
	if err != nil {
		return nil, err
	}
	if err := cursor.All(context.TODO(), &txs); err != nil {
		return nil, err
	}
	sort.Sort(txs)

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

func (s *storage) GetHeaderFromHash(blockhash string) (command.BlockHeader, error) {
	block := model.Block{}
	if err := s.db.Collection("blocks").FindOne(context.TODO(), bson.D{{Key: "hash", Value: blockhash}}).Decode(&block); err != nil {
		return command.BlockHeader{}, err
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
	blockHeader.Timestamp = time.Unix(block.Timestamp, 0)

	result, err := s.db.Collection("transactions").CountDocuments(context.TODO(), bson.D{{Key: "block_hash", Value: blockhash}})
	if err != nil {
		return command.BlockHeader{}, err
	}
	return command.BlockHeader{
		Header: blockHeader,
		Height: block.Height,
		NumTxs: result,
	}, nil
}

func (s *storage) GetHeaderFromHeight(height int32) (command.BlockHeader, error) {
	// block := &model.Block{}
	// if resp := s.db.First(block, "height = ?", height); resp.Error != nil {
	// 	return command.BlockHeader{}, resp.Error
	// }

	block := model.Block{}
	if err := s.db.Collection("blocks").FindOne(context.TODO(), bson.D{{Key: "height", Value: height}}).Decode(&block); err != nil {
		return command.BlockHeader{}, err
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
	blockHeader.Timestamp = time.Unix(block.Timestamp, 0)

	result, err := s.db.Collection("transactions").CountDocuments(context.TODO(), bson.D{{Key: "block_hash", Value: block.Hash}})
	if err != nil {
		return command.BlockHeader{}, err
	}

	return command.BlockHeader{
		Header: blockHeader,
		Height: block.Height,
		NumTxs: result,
	}, nil
}

func (s *storage) addInputsAndOutputs(txHash string, tx *wire.MsgTx) error {
	txIns := model.TxIns{}
	cursor, err := s.db.Collection("outpoints").Find(context.TODO(), bson.D{{Key: "spending_tx_hash", Value: txHash}})
	if err != nil {
		return err
	}
	if err := cursor.All(context.TODO(), &txIns); err != nil {
		return err
	}
	sort.Sort(txIns)
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

	txOuts := model.TxOuts{}
	cursor, err = s.db.Collection("outpoints").Find(context.TODO(), bson.D{{Key: "funding_tx_hash", Value: txHash}})
	if err != nil {
		return err
	}
	if err := cursor.All(context.TODO(), &txOuts); err != nil {
		return err
	}
	sort.Sort(txOuts)
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
	if err := s.db.Collection("transactions").FindOne(context.TODO(), bson.D{{Key: "hash", Value: txHash}}).Decode(&transaction); err != nil {
		return command.Transaction{}, err
	}
	block := model.Block{}
	if err := s.db.Collection("blocks").FindOne(context.TODO(), bson.D{{Key: "hash", Value: txHash}}).Decode(&block); err != nil {
		return command.Transaction{}, err
	}

	tx := wire.NewMsgTx(transaction.Version)
	tx.LockTime = transaction.LockTime
	if err := s.addInputsAndOutputs(txHash, tx); err != nil {
		return command.Transaction{}, err
	}

	if transaction.BlockHash == "" {
		return command.Transaction{
			Tx: tx,
		}, nil
	}

	return command.Transaction{
		Tx:        tx,
		BlockHash: transaction.BlockHash,
		Height:    block.Height,
		BlockTime: block.Timestamp,
	}, nil
}

func (s *storage) ListUnspent(startBlock, endBlock int, addresses []string, includeUnsafe bool, options command.ListUnspentQueryOptions) ([]model.OutPoint, error) {
	outpoints := []model.OutPoint{}
	collection := s.db.Collection("outpoints")

	if !includeUnsafe {
		resp, err := collection.Find(context.TODO(), bson.M{
			"spender": bson.M{"$in": addresses},
			"value": bson.M{
				"$gte": options.MinimumAmount,
				"$lte": options.MaximumAmount,
			},
			"fundingtx.block.height": bson.M{
				"$gte": startBlock,
				"$lte": endBlock,
			},
			"fundingtx.safe": true,
		})
		if err != nil {
			return nil, err
		}
		defer resp.Close(context.TODO())

		for resp.Next(context.TODO()) {
			var outpoint model.OutPoint
			if err := resp.Decode(&outpoint); err != nil {
				return nil, err
			}
			outpoints = append(outpoints, outpoint)
		}
	}	else {
		resp, err := collection.Find(context.TODO(), bson.M{
			"spender": bson.M{"$in": addresses},
			"value": bson.M{
				"$gte": options.MinimumAmount,
				"$lte": options.MaximumAmount,
			},
			"fundingtx.block.height": bson.M{
				"$gte": startBlock,
				"$lte": endBlock,
			},
		})
		if err != nil {
			return nil, err
		}
		defer resp.Close(context.TODO())

		for resp.Next(context.TODO()) {
			var outpoint model.OutPoint
			if err := resp.Decode(&outpoint); err != nil {
				return nil, err
			}
			outpoints = append(outpoints, outpoint)
		}
	}
	return outpoints, nil
}
