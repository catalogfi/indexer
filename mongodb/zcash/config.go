package zcash

import (
	"math/big"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

var genesisCoinbaseTx = wire.MsgTx{
	Version: 1,
	TxIn: []*wire.TxIn{
		{
			PreviousOutPoint: wire.OutPoint{
				Hash:  chainhash.Hash{},
				Index: 0xffffffff,
			},
			SignatureScript: []byte{
				0x04, 0xff, 0xff, 0x00, 0x1d, 0x01, 0x04, 0x45, /* |.......E| */
				0x4e, 0x69, 0x6e, 0x74, 0x6f, 0x6e, 0x64, 0x6f, /* |Nintondo| */
			},
			Sequence: 0xffffffff,
		},
	},
	TxOut: []*wire.TxOut{
		{
			Value: 0x00000000,
			PkScript: []byte{
				0x41, 0x04, 0x01, 0x84, 0x71, 0x0f, 0xa6, 0x89, /* |A...q...| */
				0xad, 0x50, 0x23, 0x69, 0x0c, 0x80, 0xf3, 0xa4, /* |.P#i....| */
				0x9c, 0x8f, 0x13, 0xf8, 0xd4, 0x5b, 0x8c, 0x85, /* |.....[..| */
				0x7f, 0xbc, 0xbc, 0x8b, 0xc4, 0xa8, 0xe4, 0xd3, /* |........| */
				0xeb, 0x4b, 0x10, 0xf4, 0xd4, 0x60, 0x4f, 0xa0, /* |.K...`O.| */
				0x8d, 0xce, 0x60, 0x1a, 0xaf, 0x0f, 0x47, 0x02, /* |..`...G.| */
				0x16, 0xfe, 0x1b, 0x51, 0x85, 0x0b, 0x4a, 0xcf, /* |...Q..J.| */
				0x21, 0xb1, 0x79, 0xc4, 0x50, 0x70, 0xac, 0x7b, /* |!.y.Pp.{| */
				0x03, 0xa9, 0xac, /* |...| */
			},
		},
	},
	LockTime: 0,
}

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = chainhash.Hash([chainhash.HashSize]byte{0xc4, 0xea, 0xa5, 0x88, 0x79, 0x08, 0x1d, 0xe3, 0xc2, 0x4a, 0x7b, 0x11, 0x7e, 0xd2, 0xb2, 0x83, 0x00, 0xe7, 0xec, 0x4c, 0x4c, 0x1d, 0xff, 0x1d, 0x3f, 0x12, 0x68, 0xb7, 0x85, 0x7a, 0x4d, 0xdb})
var testNet3GenesisHash = chainhash.Hash([chainhash.HashSize]byte{ // Make go vet happy.
	0x9e, 0x55, 0x50, 0x73, 0xd0, 0xc4, 0xf3, 0x64,
	0x56, 0xdb, 0x89, 0x51, 0xf4, 0x49, 0x70, 0x4d,
	0x54, 0x4d, 0x28, 0x26, 0xd9, 0xaa, 0x60, 0x63,
	0x6b, 0x40, 0x37, 0x46, 0x26, 0x78, 0x0a, 0xbb,
})
var bigOne = big.NewInt(1)

var testNet3PowLimit = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 236), bigOne)

// testNet3GenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the test network (version 3).  It is the same as the merkle root
// for the main network.
var testNet3GenesisMerkleRoot = genesisMerkleRoot

var testNet3GenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:    1,
		PrevBlock:  chainhash.Hash{},                                                   // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: testNet3GenesisMerkleRoot,                                          // 5b2a3f53f605d62c53e62932dac6925e3d74afa5a4b459745c36d42d0ed26a69
		Timestamp:  time.Unix(1477648033, 0),                                           // 2014-02-04 08:41:29 +0000 UTC
		Bits:       0x2007ffff,                                                         // 504365040 [00000ffff0000000000000000000000000000000000000000000000000000000]
		Nonce:      0x0000000000000000000000000000000000000000000000000000000000000006, // 997879
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx}, // needed
}

var ZcashTestNetParams = &chaincfg.Params{
	Name:                          "testnet",
	Net:                           wire.BitcoinNet(0xBFF91AFA),
	DefaultPort:                   "18233",
	DNSSeeds:                      []chaincfg.DNSSeed{},
	GenesisBlock:                  &testNet3GenesisBlock,
	GenesisHash:                   &testNet3GenesisHash,
	PowLimit:                      testNet3PowLimit,
	PowLimitBits:                  0x3bffffff,
	CoinbaseMaturity:              240,
	SubsidyReductionInterval:      100000,
	TargetTimespan:                time.Minute,
	TargetTimePerBlock:            time.Second * 75,
	RetargetAdjustmentFactor:      4,
	ReduceMinDifficulty:           true,
	MinDiffReductionTime:          time.Minute * 2,
	GenerateSupported:             false,
	RuleChangeActivationThreshold: 1900,
	MinerConfirmationWindow:       2000,
	Deployments: [chaincfg.DefinedDeployments]chaincfg.ConsensusDeployment{
		chaincfg.DeploymentTestDummy: {
			BitNumber: 28,
			DeploymentStarter: chaincfg.NewMedianTimeDeploymentStarter(
				time.Unix(1199145601, 0), // January 1, 2008 UTC
			),
			DeploymentEnder: chaincfg.NewMedianTimeDeploymentEnder(
				time.Unix(1230767999, 0), // December 31, 2008 UTC
			),
		},
		chaincfg.DeploymentTestDummyMinActivation: {
			BitNumber:                 22,
			CustomActivationThreshold: 1815,    // Only needs 90% hash rate.
			MinActivationHeight:       10_0000, // Can only activate after height 10k.
			DeploymentStarter: chaincfg.NewMedianTimeDeploymentStarter(
				time.Time{}, // Always available for vote
			),
			DeploymentEnder: chaincfg.NewMedianTimeDeploymentEnder(
				time.Time{}, // Never expires
			),
		},
		chaincfg.DeploymentCSV: {
			BitNumber: 0,
			DeploymentStarter: chaincfg.NewMedianTimeDeploymentStarter(
				time.Unix(1456790400, 0), // March 1st, 2016
			),
			DeploymentEnder: chaincfg.NewMedianTimeDeploymentEnder(
				time.Unix(1493596800, 0), // May 1st, 2017
			),
		},
		// chaincfg.DeploymentSegwit: {
		// 	BitNumber: 1,
		// 	DeploymentStarter: chaincfg.NewMedianTimeDeploymentStarter(
		// 		time.Unix(1462060800, 0), // May 1, 2016 UTC
		// 	),
		// },
	},

	RelayNonStdTxs: true,

	Bech32HRPSegwit: "tm",
}
