package mongodb

import (
	"context"
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
	"github.com/catalogfi/indexer/mongodb/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *storage) GetBlockLocator() (blockchain.BlockLocator, error) {
	height, err := s.GetLatestBlockHeight()
	if err != nil {
		return nil, err
	}
	locatorIDs := calculateLocator([]int{int(height)})
	blocks := []model.Block{}

	cur, err := s.db.Collection("blocks").Find(context.TODO(), bson.M{"height": bson.M{"$in": locatorIDs}})
	if err != nil {
		return nil, err
	}
	if err := cur.All(context.TODO(), &blocks); err != nil {
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

	update := bson.M{
		"$set": transaction,
	}

	txCollection := s.db.Collection("transactions")

	fOCResult, err := txCollection.UpdateOne(context.TODO(), bson.M{"hash": transactionHash}, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("error: %w , hash: %v", err, transactionHash)
	}

	if block != nil {
		transaction.BlockID = block.ID.Hex()
		transaction.BlockIndex = blockIndex
		transaction.BlockHash = block.Hash

		filter := bson.M{"hash": transactionHash}
		update := bson.M{"$set": transaction}

		_, err = txCollection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return fmt.Errorf("error: %w , filter: %v", err, filter)
		}
	}

	if fOCResult.ModifiedCount > 0 {
		// If the transaction already exists, we don't need to do anything else
		return nil
	}

	id, ok := fOCResult.UpsertedID.(primitive.ObjectID)
	if !ok {
		return fmt.Errorf("ERROR WHILE TYPECASTING")
	}
	txID := id.Hex()
	fmt.Println("putting tx", txID)

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
			result := s.db.Collection("outpoints").FindOne(context.TODO(), bson.M{"fundingtxhash": txIn.PreviousOutPoint.Hash.String(), "fundingtxindex": txIn.PreviousOutPoint.Index})
			if result.Err() != nil {
				// panic(fmt.Errorf("error %v at line %v", result.Err(), 149, txIn.PreviousOutPoint.Hash.String(), txIn.PreviousOutPoint.Index))
				return fmt.Errorf("error: %w , hash: %v", result.Err(), txIn.PreviousOutPoint.Hash.String())
			}
			err := result.Decode(&txInOut)
			if err != nil {
				return fmt.Errorf("error: %w , hash: %v", err, txIn.PreviousOutPoint.Hash.String())
			}

			txInOut.SpendingTxID = txID
			txInOut.SpendingTxHash = transactionHash
			txInOut.SpendingTxIndex = inIndex
			txInOut.Sequence = txIn.Sequence
			txInOut.SignatureScript = hex.EncodeToString(txIn.SignatureScript)
			txInOut.Witness = witnessString
			if res := s.db.Collection("outpoints").FindOneAndUpdate(context.TODO(), bson.M{"_id": txInOut.ID}, bson.M{"$set": txInOut}); res.Err() != nil {
				return fmt.Errorf("error: %w , hash: %v", res.Err(), txIn.PreviousOutPoint.Hash.String())
			}
			continue
		}

		// Create coinbase transactions
		_, err = s.db.Collection("outpoints").InsertOne(context.TODO(), &model.OutPoint{

			SpendingTxID:    txID,
			SpendingTxHash:  transactionHash,
			SpendingTxIndex: inIndex,
			Sequence:        txIn.Sequence,
			SignatureScript: hex.EncodeToString(txIn.SignatureScript),
			Witness:         witnessString,

			FundingTxHash:  txIn.PreviousOutPoint.Hash.String(),
			FundingTxIndex: txIn.PreviousOutPoint.Index,
		})
		if err != nil {
			return fmt.Errorf("error: %w , hash: %v", err, txIn.PreviousOutPoint.Hash.String())
		}
	}

	for i, txOut := range tx.TxOut {
		spenderAddress := ""

		pkScript, err := txscript.ParsePkScript(txOut.PkScript)
		if err == nil {
			addr, err := pkScript.Address(s.params)
			if err != nil {
				return fmt.Errorf("error: %w , hash: %v", err, txOut.PkScript)
			}
			spenderAddress = addr.EncodeAddress()
		}

		// Create a new outpoint
		_, err = s.db.Collection("outpoints").InsertOne(context.TODO(), &model.OutPoint{
			FundingTxID:    txID,
			FundingTxHash:  transactionHash,
			FundingTxIndex: uint32(i),
			PkScript:       hex.EncodeToString(txOut.PkScript),
			Value:          txOut.Value,
			Spender:        spenderAddress,
			Type:           pkScript.Class().String(),
		})
		if err != nil {
			return fmt.Errorf("error: %w , hash: %v", err, txOut.PkScript)
		}
	}
	return nil
}

func (s *storage) PutBlock(block *wire.MsgBlock) error {
	height := int32(-1)
	previousBlock := &model.Block{}
	// fmt.Println("check 1")
	if block.Header.PrevBlock.String() == s.params.GenesisBlock.BlockHash().String() {
		genesisBlock := btcutil.NewBlock(s.params.GenesisBlock)
		genesisBlock.SetHeight(0)

		_, err := s.db.Collection("blocks").InsertOne(context.TODO(), &model.Block{
			Hash:   genesisBlock.Hash().String(),
			Height: 0,

			IsOrphan:      false,
			PreviousBlock: genesisBlock.MsgBlock().Header.PrevBlock.String(),
			Version:       genesisBlock.MsgBlock().Header.Version,
			Nonce:         genesisBlock.MsgBlock().Header.Nonce,
			Timestamp:     genesisBlock.MsgBlock().Header.Timestamp.Unix(),
			Bits:          genesisBlock.MsgBlock().Header.Bits,
			MerkleRoot:    genesisBlock.MsgBlock().Header.MerkleRoot.String(),
		})
		if err != nil {
			return fmt.Errorf("error %v", err)
		}
		// fmt.Println("TRUEEEEE")
		_, err = s.db.Collection("transactions").InsertOne(context.TODO(), &model.Transaction{
			Hash: "0000000000000000000000000000000000000000000000000000000000000000",
		})
		if err != nil {
			return fmt.Errorf("error %v", err)
		}

		height = 1
	} else {
		// fmt.Println("check 2", block.Header.PrevBlock.String(), block.Header.Timestamp.Unix())
		resp := s.db.Collection("blocks").FindOne(context.TODO(), bson.M{"hash": block.Header.PrevBlock.String()})
		if resp.Err() != nil {
			return fmt.Errorf("failed to retrieve the stored block: %v for hash %v", resp.Err(), block.Header.PrevBlock.String())
		}
		// panic("check 3")
		err := resp.Decode(&previousBlock)
		if err != nil {
			return err
		}

		if previousBlock.IsOrphan {
			newlyOrphanedBlock := &model.Block{}
			resp := s.db.Collection("blocks").FindOne(context.TODO(), bson.M{"height": previousBlock.Height, "isorphan": false})
			if resp.Err() != nil {
				return resp.Err()
			}
			newlyOrphanedBlock.ID = primitive.NewObjectID()
			if err := resp.Decode(&newlyOrphanedBlock); err != nil {
				return err
			}

			if err := s.orphanBlock(newlyOrphanedBlock); err != nil {
				return err
			}

			previousBlock.IsOrphan = false
			_, err = s.db.Collection("blocks").UpdateByID(context.TODO(), previousBlock.ID, bson.M{"$set": previousBlock})
			if err != nil {
				return err
			}
		}

		height = previousBlock.Height + 1
	}

	blockAtHeight := &model.Block{}

	// fmt.Println("check 4" + strconv.Itoa(int(height)))
	err := s.db.Collection("blocks").FindOne(context.TODO(), bson.M{"height": height}).Decode(&blockAtHeight)
	if err != mongo.ErrNoDocuments {
		if err != nil {
			// panic(fmt.Errorf("error %v line %v", err, 296))
			return fmt.Errorf("failed to retrieve the stored block: %v", err)
		} else {
			if err := s.orphanBlock(blockAtHeight); err != nil {
				// panic(fmt.Errorf("error %v line %v", err, 301))
				return err
			}
		}
		// return fmt.Errorf("failed ++ retrieve the stored block: %v", err)
	}

	bblock := &model.Block{
		Hash:   block.Header.BlockHash().String(),
		Height: height,

		IsOrphan:      false,
		PreviousBlock: block.Header.PrevBlock.String(),
		Version:       block.Header.Version,
		Nonce:         block.Header.Nonce,
		Timestamp:     block.Header.Timestamp.Unix(),
		Bits:          block.Header.Bits,
		MerkleRoot:    block.Header.MerkleRoot.String(),
	}
	_, err = s.db.Collection("blocks").InsertOne(context.TODO(), bblock)
	if err != nil {
		return fmt.Errorf("InsertOne error: %v", err)
	}
	// fmt.Println("Block", bblock.Height, "has been added to the database", bblock)
	fmt.Println("")
	for i, tx := range block.Transactions {
		if err := s.putTx(tx, bblock, uint32(i)); err != nil {
			return fmt.Errorf("InsertOne error: %v", err)
		}
	}

	aBlock := &model.Block{}

	err = s.db.Collection("blocks").FindOne(context.TODO(), bson.M{"hash": block.Header.BlockHash().String()}).Decode(&aBlock)
	if err != nil {
		return fmt.Errorf("failed to retrieve the stored block: %v", err)
	}
	fmt.Println("Block", aBlock.Height, "has been added to the database", aBlock)

	return nil
}

func (s *storage) orphanBlock(block *model.Block) error {
	block.IsOrphan = true
	cur, err := s.db.Collection("transactions").Find(context.TODO(), bson.D{{"BlockHash", block.Hash}})
	if err != nil {
		return err
	}
	txs := []model.Transaction{}
	if err := cur.All(context.Background(), &txs); err != nil {
		return err
	}
	_, err = s.db.Collection("transactions").UpdateMany(context.TODO(), bson.D{{"BlockHash", block.Hash}}, bson.M{"$set": bson.M{"BlockID": 0, "BlockHash": "", "BlockIndex": 0}})
	if err != nil {
		return err
	}
	_, err = s.db.Collection("blocks").UpdateOne(context.Background(), bson.D{{Key: "_id", Value: block.ID}}, bson.M{"$set": block}, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) Params() *chaincfg.Params {
	return s.params
}
