package store_test

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/catalogfi/indexer/model"
	"github.com/catalogfi/indexer/store"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var counter = 0

func TestMyPackage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MyPackage Suite")
}
func GetBlockByHeight(height int32, db *gorm.DB) (*model.Block, error) {
	block := &model.Block{}
	resp := db.First(block, "height = ?", height)
	return block, resp.Error
}

var _ = Describe("Tests", Ordered, func() {

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

			// println(h)
			s.PutBlock(block, "bitcoin")
			// println(h + 1)

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
			counter++
			println("counter value is 6 : ")
			println(counter)
			db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
			if err != nil {
				panic(err)
			}
			s := store.NewStorage(&chaincfg.RegressionNetParams, db)
			h, err := s.GetLatestBlockHeight()
			Expect(err).To(BeNil())
			prevblock := &model.Block{}
			resp := db.First(prevblock, "height = ?", h)
			println(h)
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

			println(h)
			s.PutBlock(block, "bitcoin")
			println(h + 1)

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
			// h, err := s.GetLatestUnorphanBlockHeight()
			Expect(err).To(BeNil())
			prevblock := &model.Block{}

			heightOfPrevBlock := h
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

			println(h)
			err = s.PutBlock(block, "bitcoin")
			println(h + 1)
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

	var _ = Describe("GetHeaderFromHeight", func() {

		It("should get the blocks header given the height", func() {
			counter++
			println("counter value is 1 : ")
			println(counter)
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

			merkleRootHash, err := chainhash.NewHashFromStr("ae1f680b35dae02d3eadba15b0644740c09263972589a0953c1af4665d1a681a")
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
			// println(h)

			println(h)
			s.PutBlock(block, "bitcoin")
			println(h + 1)

			header, err := s.GetHeaderFromHeight(h + 1)
			// println(header.Height)
			Expect(err).To(BeNil())
			// Expect(header).To(Equal(block.Header))
			Expect(header.Header.Version).To(Equal(block.Header.Version))
			Expect(header.Header.PrevBlock).To(Equal(block.Header.PrevBlock))
			Expect(header.Header.MerkleRoot).To(Equal(block.Header.MerkleRoot))
			Expect(header.Header.Bits).To(Equal(block.Header.Bits))
			Expect(header.Header.Nonce).To(Equal(block.Header.Nonce))
			Expect(header.Height).To(Equal(h + 1))
		})
	})

	var _ = Describe("GetHeaderFromHash", func() {

		It("should get the blocks header given the hash", func() {
			counter++
			println("counter value is 2 : ")
			println(counter)
			db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
			if err != nil {
				panic(err)
			}
			s := store.NewStorage(&chaincfg.RegressionNetParams, db)
			// h, err := s.GetLatestBlockHeight()
			// Expect(err).To(BeNil())
			prevblock := &model.Block{}
			resp := db.Order("height desc").First(prevblock)
			Expect(resp.Error).To(BeNil())
			h := prevblock.Height
			prevBlockHash, err := chainhash.NewHashFromStr(prevblock.Hash)
			Expect(err).To(BeNil())

			merkleRootHash, err := chainhash.NewHashFromStr("ad1f680b35dae02d3eadba15b0644740c09263972589a0953c1af4665d1a681a")
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

			println(h)
			s.PutBlock(block, "bitcoin")
			println(h + 1)

			// Expect(db.Create(block).Error).To(BeNil())
			header, err := s.GetHeaderFromHash(block.Header.BlockHash().String())
			Expect(err).To(BeNil())
			Expect(header.Header.Version).To(Equal(block.Header.Version))
			Expect(header.Header.PrevBlock).To(Equal(block.Header.PrevBlock))
			Expect(header.Header.MerkleRoot).To(Equal(block.Header.MerkleRoot))
			Expect(header.Header.Bits).To(Equal(block.Header.Bits))
			Expect(header.Header.Nonce).To(Equal(block.Header.Nonce))
			Expect(header.Height).To(Equal(h + 1))
		})
	})

	var _ = Describe("GetBlockFromHash", func() {

		It("should get the blocks given the hash", func() {
			counter++
			println("counter value is 3 : ")
			println(counter)
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

			merkleRootHash, err := chainhash.NewHashFromStr("ad1f680b35dae02d3eadba15b0644740c09263972589a0953c1af4665d1a681a")
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

			println(h)
			s.PutBlock(block, "bitcoin")
			println(h + 1)

			// Expect(db.Create(block).Error).To(BeNil())
			btcblock1, err := s.GetBlockFromHash(block.Header.BlockHash().String())
			Expect(err).To(BeNil())
			Expect(btcblock1.MsgBlock().Header.Version).To(Equal(block.Header.Version))
			Expect(btcblock1.MsgBlock().Header.MerkleRoot).To(Equal(block.Header.MerkleRoot))
			Expect(btcblock1.MsgBlock().Header.Bits).To(Equal(block.Header.Bits))
			Expect(btcblock1.MsgBlock().Header.Nonce).To(Equal(block.Header.Nonce))
			Expect(btcblock1.MsgBlock().Header.PrevBlock).To(Equal(block.Header.PrevBlock))
		})
	})

	var _ = Describe("GetBlockHash", func() {

		It("should get the blocks hash given the height", func() {
			counter++
			println("counter value is 4 : ")
			println(counter)
			db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
			if err != nil {
				panic(err)
			}
			s := store.NewStorage(&chaincfg.RegressionNetParams, db)
			// prevblock := &model.Block{}
			// h, err := s.GetLatestBlockHeight()
			// Expect(err).To(BeNil())
			prevblock := &model.Block{}
			resp := db.Order("height desc").First(prevblock)
			// resp := db.First(prevblock, "height = ?", h)
			// println(h)
			h := prevblock.Height
			Expect(resp.Error).To(BeNil())

			prevBlockHash, err := chainhash.NewHashFromStr(prevblock.Hash)
			Expect(err).To(BeNil())

			merkleRootHash, err := chainhash.NewHashFromStr("5b2a3f53f605d62c53e62932dac6925e3d74afa5a4b459745c36d42d0ed26a69")
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

			println(h)
			s.PutBlock(block, "bitcoin")
			println(h + 1)
			hash, err := s.GetBlockHash(h + 1)
			Expect(err).To(BeNil())
			Expect(hash).To(Equal(block.Header.BlockHash().String()))
		})
	})

	var _ = Describe("GetLatestBlockHash", func() {

		It("should get the blocks hash with the max height", func() {
			counter++
			println("counter value is 5 : ")
			println(counter)
			db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
			if err != nil {
				panic(err)
			}
			s := store.NewStorage(&chaincfg.RegressionNetParams, db)
			prevblock := &model.Block{}
			h, err := s.GetLatestBlockHeight()
			Expect(err).To(BeNil())
			resp := db.First(prevblock, "height = ?", h)
			println(h)
			Expect(resp.Error).To(BeNil())

			prevBlockHash, err := chainhash.NewHashFromStr(prevblock.Hash)
			Expect(err).To(BeNil())

			merkleRootHash, err := chainhash.NewHashFromStr("5b2a3f53f605d62c53e62932dac6925e3d74afa5a4b459745c36d42d0ed26a69")
			Expect(err).To(BeNil())
			rand.Seed(time.Now().UnixNano())
			block := &wire.MsgBlock{
				Header: wire.BlockHeader{
					Version:    1,
					PrevBlock:  *prevBlockHash,
					MerkleRoot: *merkleRootHash,
					Timestamp:  time.Unix(1231006505, 0), // Set a specific timestamp
					Bits:       rand.Uint32(),
					Nonce:      2083236893,
				},
				Transactions: []*wire.MsgTx{},
			}

			//Another way for testing for Hash by hardcoding the hash value
			bblock := &model.Block{
				Hash:   block.Header.BlockHash().String(),
				Height: h + 1,

				IsOrphan:      false,
				PreviousBlock: block.Header.PrevBlock.String(),
				Version:       block.Header.Version,
				Nonce:         block.Header.Nonce,
				Timestamp:     block.Header.Timestamp,
				Bits:          block.Header.Bits,
				MerkleRoot:    block.Header.MerkleRoot.String(),
			}
			result := db.Create(bblock)
			Expect(result.Error).To(BeNil())
			hash1, err := s.GetLatestBlockHash()
			Expect(err).To(BeNil())
			Expect(hash1).To(Equal(bblock.Hash))
		})
	})

	var _ = Describe("CalculateLocator", func() {

		It("should return an empty locator for an empty input", func() {
			counter++
			println("counter value is 7: ")
			println(counter)

			loc := []int{}
			locator := store.CalculateLocator(loc)
			Expect(locator).To(BeEmpty())
		})

		It("should return the same locator if the height is already 0", func() {
			loc := []int{0, 0, 0}
			locator := store.CalculateLocator(loc)
			Expect(locator).To(Equal(loc))
		})

		It("should calculate the locator correctly", func() {
			loc := []int{10}
			locator := store.CalculateLocator(loc)
			Expect(locator).To(Equal([]int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0}))
		})

		It("should handle large locator sizes", func() {
			loc := make([]int, int(math.Pow(2, 12))+1)
			locator := store.CalculateLocator(loc)
			Expect(len(locator)).To(Equal(4097))
			Expect(locator[12]).To(Equal(0))
		})
	})

	var _ = Describe("GetBlockLocator", func() {

		It("should return the correct block locator", func() {
			counter++
			println("counter value is 8 : ")
			println(counter)
			db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
			if err != nil {
				panic(err)
			}
			s := store.NewStorage(&chaincfg.RegressionNetParams, db)
			h, err := s.GetLatestBlockHeight()
			Expect(err).To(BeNil())
			prevblock := &model.Block{}
			resp := db.First(prevblock, "height = ?", h)
			// println(height)
			Expect(resp.Error).To(BeNil())

			prevBlockHash, err := chainhash.NewHashFromStr(prevblock.Hash)
			Expect(err).To(BeNil())

			merkleRootHash, err := chainhash.NewHashFromStr("5a2a3f53f605d62c53e62932dac6925e3d74afa5a4b459745c36d42d0ed26a69")
			Expect(err).To(BeNil())

			block := &wire.MsgBlock{
				Header: wire.BlockHeader{
					Version:    1,
					PrevBlock:  *prevBlockHash,
					MerkleRoot: *merkleRootHash,
					Timestamp:  time.Unix(1231006505, 0), // Set a specific timestamp
					Bits:       486604799,
					Nonce:      8,
				},
				Transactions: []*wire.MsgTx{},
			}
			println(h)
			s.PutBlock(block, "bitcoin")
			println(h + 1)
			//This creates a block in the database our idea is to check whether the last added block is same as the one we added
			// Expect(db.Create(block).Error).To(BeNil())
			locator, err := s.GetBlockLocator()
			Expect(err).To(BeNil())
			// Expect(len(locator)).To(Equal(22))
			Expect(locator[0].String()).To(Equal(block.Header.BlockHash().String()))
		})

	})

	var _ = Describe("GetLatestBlockHeight", func() {

		It("should return the correct block height", func() {
			counter++
			println("counter value is 9 : ")
			println(counter)
			db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
			if err != nil {
				panic(err)
			}
			s := store.NewStorage(&chaincfg.RegressionNetParams, db)
			// height, err := s.GetLatestBlockHeight()
			// Expect(err).To(BeNil())
			prevblock := &model.Block{}
			resp := db.Order("height desc").First(prevblock)
			h := prevblock.Height
			// println(height)
			Expect(resp.Error).To(BeNil())

			prevBlockHash, err := chainhash.NewHashFromStr(prevblock.Hash)
			Expect(err).To(BeNil())

			merkleRootHash, err := chainhash.NewHashFromStr("5a2a3f53f605d62c53e62932dac6925e3d74afa5a4b459745c36d42d0ed26a69")
			Expect(err).To(BeNil())

			block := &wire.MsgBlock{
				Header: wire.BlockHeader{
					Version:    1,
					PrevBlock:  *prevBlockHash,
					MerkleRoot: *merkleRootHash,
					Timestamp:  time.Unix(1231006505, 0), // Set a specific timestamp
					Bits:       486604799,
					Nonce:      8,
				},
				Transactions: []*wire.MsgTx{},
			}
			println(h)
			s.PutBlock(block, "bitcoin")
			println(h + 1)
			newHeight, err := s.GetLatestBlockHeight()
			Expect(err).To(BeNil())
			Expect(newHeight).To(Equal(h + 1))
		})
	})

})

//Not a test case using for debugging pursoses.
// var _ = Describe("NewPeer",Serial , func() {
// 	It("should create a new outbound peer",Serial , func() {
// 		db, err := model.NewDB(sqlite.Open("gorm.db"), &gorm.Config{})
// 		if err != nil {
// 			panic(err)
// 		}
// 		s := store.NewStorage(&chaincfg.RegressionNetParams, db)
// 		h, err := s.GetLatestBlockHeight()
// 		Expect(err).To(BeNil())
// 		prevblock := &model.Block{}
// 		resp := db.First(prevblock, "height = ?", h)
// 		println(h)
// 		Expect(resp.Error).To(BeNil())

// 		prevBlockHash, err := chainhash.NewHashFromStr(prevblock.Hash)
// 		Expect(err).To(BeNil())

// 		merkleRootHash, err := chainhash.NewHashFromStr("5b2a3f53f605d62c53e62932dac6925e3d74afa5a4b459745c36d42d0ed26a69")
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
// 		bblock := &model.Block{
// 			Hash:   "3d2160a3b5dc4a9d62e7e66a295f70313ac808440ef7400d6c0772171ce973a5",
// 			Height: h + 1,

// 			IsOrphan:      false,
// 			PreviousBlock: block.Header.PrevBlock.String(),
// 			Version:       block.Header.Version,
// 			Nonce:         block.Header.Nonce,
// 			Timestamp:     block.Header.Timestamp,
// 			Bits:          block.Header.Bits,
// 			MerkleRoot:    block.Header.MerkleRoot.String(),
// 		}
// 		if result := db.Create(bblock); result.Error != nil {
// 			println(result.Error)
// 			return
// 		}
// 	})
// })
