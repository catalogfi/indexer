package store_test

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/catalogfi/indexer/model"
	"github.com/catalogfi/indexer/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMyPackage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MyPackage Suite")
}

// var _ = Describe("GetHeaderFromHeight", func() {
// 	It("should get the blocks header given the height", func() {
// 		db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
// 		if err != nil {
// 			panic(err)
// 		}
// 		s := store.NewStorage(&chaincfg.RegressionNetParams, db)
// 		h, err := s.GetLatestBlockHeight()
// 		Expect(err).To(BeNil())
// 		prevblock := &model.Block{}
// 		resp := db.First(prevblock, "height = ?", h)
// 		Expect(resp.Error).To(BeNil())
// 		prevBlockHash, err := chainhash.NewHashFromStr(prevblock.Hash)
// 		Expect(err).To(BeNil())

// 		merkleRootHash, err := chainhash.NewHashFromStr("ae1f680b35dae02d3eadba15b0644740c09263972589a0953c1af4665d1a681a")
// 		Expect(err).To(BeNil())
// 		block := &wire.MsgBlock{
// 			Header: wire.BlockHeader{
// 				Version:    1,
// 				PrevBlock:  *prevBlockHash,
// 				MerkleRoot: *merkleRootHash,
// 				Timestamp:  time.Unix(1231006505, 0), // Set a specific timestamp
// 				Bits:       486604799,
// 				Nonce:      2083236893,
// 			},
// 			Transactions: []*wire.MsgTx{},
// 		}

// 		s.PutBlock(block)

// 		header, err := s.GetHeaderFromHeight(h + 1)
// 		Expect(err).To(BeNil())
// 		// Expect(header).To(Equal(block.Header))
// 		Expect(header.Header.Version).To(Equal(block.Header.Version))
// 		Expect(header.Header.PrevBlock).To(Equal(block.Header.PrevBlock))
// 		Expect(header.Header.MerkleRoot).To(Equal(block.Header.MerkleRoot))
// 		Expect(header.Header.Bits).To(Equal(block.Header.Bits))
// 		Expect(header.Header.Nonce).To(Equal(block.Header.Nonce))
// 		Expect(header.Height).To(Equal(h + 1))
// 	})
// })

func GetBlockByHeight(height int32, db *gorm.DB) (*model.Block, error) {
	block := &model.Block{}
	resp := db.First(block, "height = ?", height)
	return block, resp.Error
}

// var _ = Describe("GetHeaderFromHash", func() {
// 	It("should get the blocks header given the hash", func() {
// 		db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
// 		if err != nil {
// 			panic(err)
// 		}
// 		s := store.NewStorage(&chaincfg.RegressionNetParams, db)
// 		h, err := s.GetLatestBlockHeight()
// 		Expect(err).To(BeNil())
// 		prevblock := &model.Block{}
// 		resp := db.First(prevblock, "height = ?", h)
// 		Expect(resp.Error).To(BeNil())
// 		prevBlockHash, err := chainhash.NewHashFromStr(prevblock.Hash)
// 		Expect(err).To(BeNil())

// 		merkleRootHash, err := chainhash.NewHashFromStr("ad1f680b35dae02d3eadba15b0644740c09263972589a0953c1af4665d1a681a")
// 		Expect(err).To(BeNil())
// 		block := &wire.MsgBlock{
// 			Header: wire.BlockHeader{
// 				Version:    1,
// 				PrevBlock:  *prevBlockHash,
// 				MerkleRoot: *merkleRootHash,
// 				Timestamp:  time.Unix(1231006505, 0), // Set a specific timestamp
// 				Bits:       486604799,
// 				Nonce:      2083236893,
// 			},
// 			Transactions: []*wire.MsgTx{},
// 		}

// 		s.PutBlock(block)

// 		// Expect(db.Create(block).Error).To(BeNil())
// 		header, err := s.GetHeaderFromHash(block.Header.BlockHash().String())
// 		Expect(err).To(BeNil())
// 		Expect(header.Header.Version).To(Equal(block.Header.Version))
// 		Expect(header.Header.PrevBlock).To(Equal(block.Header.PrevBlock))
// 		Expect(header.Header.MerkleRoot).To(Equal(block.Header.MerkleRoot))
// 		Expect(header.Header.Bits).To(Equal(block.Header.Bits))
// 		Expect(header.Header.Nonce).To(Equal(block.Header.Nonce))
// 		Expect(header.Height).To(Equal(h + 1))
// 		// Expect(header.MsgBlock().Header.Version).To(Equal(block.Header.Version))
// 	})
// })

// var _ = Describe("GetBlockFromHash", func() {
// 	It("should get the blocks given the hash", func() {
// 		db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
// 		if err != nil {
// 			panic(err)
// 		}
// 		s := store.NewStorage(&chaincfg.RegressionNetParams, db)
// 		h, err := s.GetLatestBlockHeight()
// 		Expect(err).To(BeNil())
// 		prevblock := &model.Block{}
// 		resp := db.First(prevblock, "height = ?", h)
// 		Expect(resp.Error).To(BeNil())
// 		prevBlockHash, err := chainhash.NewHashFromStr(prevblock.Hash)
// 		Expect(err).To(BeNil())

// 		merkleRootHash, err := chainhash.NewHashFromStr("ad1f680b35dae02d3eadba15b0644740c09263972589a0953c1af4665d1a681a")
// 		Expect(err).To(BeNil())
// 		block := &wire.MsgBlock{
// 			Header: wire.BlockHeader{
// 				Version:    1,
// 				PrevBlock:  *prevBlockHash,
// 				MerkleRoot: *merkleRootHash,
// 				Timestamp:  time.Unix(1231006505, 0), // Set a specific timestamp
// 				Bits:       486604799,
// 				Nonce:      2083236893,
// 			},
// 			Transactions: []*wire.MsgTx{},
// 		}

// 		s.PutBlock(block)

// 		// Expect(db.Create(block).Error).To(BeNil())
// 		btcblock1, err := s.GetBlockFromHash(block.Header.BlockHash().String())
// 		Expect(err).To(BeNil())
// 		Expect(btcblock1.MsgBlock().Header.Version).To(Equal(block.Header.Version))
// 		Expect(btcblock1.MsgBlock().Header.MerkleRoot).To(Equal(block.Header.MerkleRoot))
// 		Expect(btcblock1.MsgBlock().Header.Bits).To(Equal(block.Header.Bits))
// 		Expect(btcblock1.MsgBlock().Header.Nonce).To(Equal(block.Header.Nonce))
// 		// Expect(btcblock1.MsgBlock().Header.Timestamp).To(Equal(block.Header.Timestamp))
// 		Expect(btcblock1.MsgBlock().Header.PrevBlock).To(Equal(block.Header.PrevBlock))
// 	})
// })

// var _ = Describe("GetBlockHash", func() {

// 	It("should get the blocks hash given the height", func() {
// 		db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
// 		if err != nil {
// 			panic(err)
// 		}
// 		s := store.NewStorage(&chaincfg.RegressionNetParams, db)
// 		block1 := &model.Block{
// 			Hash:   "2c67b822e4a755c81dc274da32d968e815f89e084614b9613daca60ba1e8411c",
// 			Height: 500,
// 		}
// 		block2 := &model.Block{
// 			Hash:   "abcd",
// 			Height: 1000,
// 		}
// 		block3 := &model.Block{
// 			Hash:   "1234",
// 			Height: 1500,
// 		}

// 		Expect(db.Create(block1).Error).To(BeNil())
// 		hash1, err := s.GetBlockHash(500)
// 		Expect(err).To(BeNil())
// 		Expect(hash1).To(Equal(block1.Hash))

// 		Expect(db.Create(block2).Error).To(BeNil())
// 		hash2, err := s.GetBlockHash(1000)
// 		Expect(err).To(BeNil())
// 		Expect(hash2).To(Equal(block2.Hash))

// 		Expect(db.Create(block3).Error).To(BeNil())
// 		hash3, err := s.GetBlockHash(1500)
// 		Expect(err).To(BeNil())
// 		Expect(hash3).To(Equal(block3.Hash))
// 	})
// })

// var _ = Describe("GetLatestBlockHash", func() {

// 	It("should get the blocks hash with the max height", func() {
// 		db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
// 		if err != nil {
// 			panic(err)
// 		}
// 		s := store.NewStorage(&chaincfg.RegressionNetParams, db)
// 		block1 := &model.Block{
// 			Hash:   "2c67b822e4a755c81dc274da32d968e815f89e084614b9613daca60ba1e8411c",
// 			Height: 2000,
// 		}
// 		block2 := &model.Block{
// 			Hash:   "abcd",
// 			Height: 10000,
// 		}
// 		block3 := &model.Block{
// 			Hash:   "1234",
// 			Height: 15000,
// 		}
// 		Expect(db.Create(block1).Error).To(BeNil())
// 		hash1, err := s.GetLatestBlockHash()
// 		Expect(err).To(BeNil())
// 		Expect(hash1).To(Equal(block1.Hash))

// 		Expect(db.Create(block2).Error).To(BeNil())
// 		hash2, err := s.GetLatestBlockHash()
// 		Expect(err).To(BeNil())
// 		Expect(hash2).To(Equal(block2.Hash))

// 		Expect(db.Create(block3).Error).To(BeNil())
// 		hash3, err := s.GetLatestBlockHash()
// 		Expect(err).To(BeNil())
// 		Expect(hash3).To(Equal(block3.Hash))
// 	})
// })

var _ = Describe("PutBlock", func() {

	It("should insert a genesis block to the database", func() {
		db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		s := store.NewStorage(&chaincfg.RegressionNetParams, db)
		block := &wire.MsgBlock{
			Header: wire.BlockHeader{
				Version:    1,
				PrevBlock:  *chaincfg.RegressionNetParams.GenesisHash,
				MerkleRoot: *chaincfg.RegressionNetParams.GenesisHash,
				Timestamp:  time.Unix(1231006505, 0), // Set a specific timestamp
				Bits:       486604799,
				Nonce:      2083236893,
			},
			Transactions: []*wire.MsgTx{},
		}

		s.PutBlock(block)

		block1 := &model.Block{}
		resp := db.First(block1, "height = ?", 0)
		Expect(resp.Error).To(BeNil())
		Expect(err).To(BeNil())
		Expect(block.Header.PrevBlock.String()).To(Equal(block1.Hash))
		// Expect(block1.Hash).To(Equal(block.Header.BlockHash().String()))
		Expect(block1.Height).To(Equal(int32(0)))
		Expect(block1.IsOrphan).To(BeFalse())
		// Expect(block1.PreviousBlock).To(Equal(block.Header.PrevBlock.String()))
		Expect(block1.Version).To(Equal(block.Header.Version))
		// Expect(block1.Nonce).To(Equal(block.Header.Nonce))
		// Expect(block1.Timestamp).To(Equal(block.Header.Timestamp))
		// Expect(block1.Bits).To(Equal(block.Header.Bits))
		// Expect(block1.MerkleRoot).To(Equal(block.Header.MerkleRoot.String()))

	})

	It("should insert a new block at the end of the database", func() {
		db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		s := store.NewStorage(&chaincfg.RegressionNetParams, db)
		h, err := s.GetLatestBlockHeight()
		Expect(err).To(BeNil())
		prevblock := &model.Block{}
		resp := db.First(prevblock, "height = ?", h)
		Expect(resp.Error).To(BeNil())

		prevBlockHash, err := chainhash.NewHashFromStr(prevblock.Hash)
		Expect(err).To(BeNil())

		merkleRootHash, err := chainhash.NewHashFromStr("ac1f680b35dae02d3eadba15b0644740c09263972589a0953c1af4665d1a681a")
		Expect(err).To(BeNil())

		block := &wire.MsgBlock{
			Header: wire.BlockHeader{
				Version:    1,
				PrevBlock:  *prevBlockHash,
				MerkleRoot: *merkleRootHash,
				Timestamp:  time.Unix(1231006505, 0), // Set a specific timestamp
				Bits:       486604799,
				Nonce:      2083236893,
			},
			Transactions: []*wire.MsgTx{},
		}

		s.PutBlock(block)

		nh, err := s.GetLatestBlockHeight()

		// Hash, err := chainhash.NewHashFromStr(block1.Hash)
		Expect(err).To(BeNil())
		// Expect(block.Header.PrevBlock.String()).To(Equal(Hash.String()))
		Expect(nh).To(Equal(h + 1))

		checkBlock := &model.Block{}
		r := db.First(checkBlock, "height = ?", nh)
		Expect(r.Error).To(BeNil())
		Expect(checkBlock.Hash).To(Equal(block.Header.BlockHash().String()))
		Expect(checkBlock.Hash).To(Equal(block.Header.BlockHash().String()))
		Expect(checkBlock.Height).To(Equal(h + 1))
		Expect(checkBlock.IsOrphan).To(BeFalse())
		Expect(checkBlock.PreviousBlock).To(Equal(block.Header.PrevBlock.String()))
		Expect(checkBlock.Version).To(Equal(block.Header.Version))
		Expect(checkBlock.Nonce).To(Equal(block.Header.Nonce))
		// Expect(checkBlock.Timestamp).To(Equal(block.Header.Timestamp))
		Expect(checkBlock.Bits).To(Equal(block.Header.Bits))
	})

	It("should insert an orphan block correctly", func() {
		db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
		Expect(err).To(BeNil())
		s := store.NewStorage(&chaincfg.RegressionNetParams, db)

		h, err := s.GetLatestBlockHeight()
		Expect(err).To(BeNil())
		prevblock := &model.Block{}

		heightOfPrevBlock := h - 5
		resp := db.First(prevblock, "height = ?", heightOfPrevBlock)
		Expect(resp.Error).To(BeNil())
		prevBlockHash, err := chainhash.NewHashFromStr(prevblock.Hash)
		Expect(err).To(BeNil())

		merkleRootHash, err := chainhash.NewHashFromStr("af1f680b35dae02d3eadba15b0644740c09263972589a0953c1af4665d1a682a")
		Expect(err).To(BeNil())

		// Create the orphan block
		block := &wire.MsgBlock{
			Header: wire.BlockHeader{
				Version:    1,
				PrevBlock:  *prevBlockHash,
				MerkleRoot: *merkleRootHash,
				Timestamp:  time.Unix(123456789, 0),
				Bits:       486604799,
				Nonce:      2083236893,
			},
			Transactions: []*wire.MsgTx{},
		}

		err = s.PutBlock(block)
		Expect(err).To(BeNil())

		blockFromDB := &model.Block{}
		resp1 := db.First(blockFromDB, "height = ? AND is_orphan = ?", heightOfPrevBlock+1, false)

		Expect(resp1.Error).To(BeNil())
		Expect(err).To(BeNil())
		Expect(blockFromDB.IsOrphan).To(BeFalse())
		Expect(blockFromDB.Hash).To(Equal(block.Header.BlockHash().String()))
		Expect(blockFromDB.Height).To(Equal(heightOfPrevBlock + 1))
		Expect(blockFromDB.PreviousBlock).To(Equal(block.Header.PrevBlock.String()))
		Expect(blockFromDB.Version).To(Equal(block.Header.Version))
		Expect(blockFromDB.Nonce).To(Equal(block.Header.Nonce))
		Expect(blockFromDB.Bits).To(Equal(block.Header.Bits))
		Expect(blockFromDB.MerkleRoot).To(Equal(block.Header.MerkleRoot.String()))
	})
})

// var _ = Describe("CalculateLocator", func() {
// 	It("should return an empty locator for an empty input", func() {
// 		loc := []int{}
// 		locator := store.CalculateLocator(loc)
// 		Expect(locator).To(BeEmpty())
// 	})

// 	It("should return the same locator if the height is already 0", func() {
// 		loc := []int{0, 0, 0}
// 		locator := store.CalculateLocator(loc)
// 		Expect(locator).To(Equal(loc))
// 	})

// 	It("should calculate the locator correctly", func() {
// 		loc := []int{10}
// 		locator := store.CalculateLocator(loc)
// 		Expect(locator).To(Equal([]int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0}))
// 	})

// 	It("should handle large locator sizes", func() {
// 		loc := make([]int, int(math.Pow(2, 12))+1)
// 		locator := store.CalculateLocator(loc)
// 		Expect(len(locator)).To(Equal(4097))
// 		Expect(locator[12]).To(Equal(0))
// 	})
// })

// var _ = Describe("GetBlockLocator", func() {
// 	It("should return the correct block locator", func() {

// 		db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
// 		if err != nil {
// 			panic(err)
// 		}
// 		str := store.NewStorage(&chaincfg.RegressionNetParams, db)
// 		h, err := str.GetLatestBlockHeight()
// 		Expect(err).To(BeNil())
// 		block := &model.Block{
// 			Hash:   "2c67b822e4a755c81dc274da32d968e815f89e084614b9613daca60ba1e8411c",
// 			Height: h,
// 		}
// 		//This creates a block in the database our idea is to check whether the last added block is same as the one we added
// 		Expect(db.Create(block).Error).To(BeNil())
// 		locator, err := str.GetBlockLocator()
// 		Expect(err).To(BeNil())
// 		// Expect(len(locator)).To(Equal(22))
// 		Expect(locator[0].String()).To(Equal(block.Hash))
// 	})

// })

// var _ = Describe("GetLatestBlockHeight", func() {
// 	It("should return the correct block height", func() {
// 		db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
// 		if err != nil {
// 			panic(err)
// 		}
// 		s := store.NewStorage(&chaincfg.RegressionNetParams, db)
// 		height, err := s.GetLatestBlockHeight()
// 		Expect(err).To(BeNil())
// 		// Expect(height).To(Equal(int32(165)))
// 		block := &model.Block{
// 			Hash:   "2c67b822e4a755c81dc274da32d968e815f89e084614b9613daca60ba1e1111c",
// 			Height: height + 1,
// 		}
// 		Expect(db.Create(block).Error).To(BeNil())
// 		newHeight, err := s.GetLatestBlockHeight()
// 		Expect(err).To(BeNil())
// 		Expect(newHeight).To(Equal(height + 1))
// 	})
// })
