package dogecoin

import (
	"math/big"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

func newHashFromStr(hexStr string) *chainhash.Hash {
	hash, err := chainhash.NewHashFromStr(hexStr)
	if err != nil {
		panic(err)
	}
	return hash
}

const (
	// MainNet represents the main dogecoin network.
	MainNet wire.BitcoinNet = 0xc0c0c0c0

	// TestNet3 represents the test network (version 3).
	TestNet3 wire.BitcoinNet = 0xdcb7c1fc
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
			Value: 0x12a05f200,
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

// genesisHash is the hash of the first block in the block chain for the main
// network (genesis block).
var genesisHash = chainhash.Hash([chainhash.HashSize]byte{ // Make go vet happy.
	0x91, 0x56, 0x35, 0x2c, 0x18, 0x18, 0xb3, 0x2e,
	0x90, 0xc9, 0xe7, 0x92, 0xef, 0xd6, 0xa1, 0x1a,
	0x82, 0xfe, 0x79, 0x56, 0xa6, 0x30, 0xf0, 0x3b,
	0xbe, 0xe2, 0x36, 0xce, 0xda, 0xe3, 0x91, 0x1a,
})

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = chainhash.Hash([chainhash.HashSize]byte{ // Make go vet happy.
	0x69, 0x6a, 0xd2, 0x0e, 0x2d, 0xd4, 0x36, 0x5c,
	0x74, 0x59, 0xb4, 0xa4, 0xa5, 0xaf, 0x74, 0x3d,
	0x5e, 0x92, 0xc6, 0xda, 0x32, 0x29, 0xe6, 0x53,
	0x2c, 0xd6, 0x05, 0xf6, 0x53, 0x3f, 0x2a, 0x5b,
})

// genesisBlock defines the genesis block of the block chain which serves as the
// public transaction ledger for the main network.
var genesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:    1,
		PrevBlock:  chainhash.Hash{},         // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: genesisMerkleRoot,        // 5b2a3f53f605d62c53e62932dac6925e3d74afa5a4b459745c36d42d0ed26a69
		Timestamp:  time.Unix(1386325540, 0), // 2013-12-06 10:25:40 +0000 UTC
		Bits:       0x1e0ffff0,               // 504365040 [00000ffff0000000000000000000000000000000000000000000000000000000]
		Nonce:      0x18667,                  // 99943
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// testNet3GenesisHash is the hash of the first block in the block chain for the
// test network (version 3).
var testNet3GenesisHash = chainhash.Hash([chainhash.HashSize]byte{ // Make go vet happy.
	0x9e, 0x55, 0x50, 0x73, 0xd0, 0xc4, 0xf3, 0x64,
	0x56, 0xdb, 0x89, 0x51, 0xf4, 0x49, 0x70, 0x4d,
	0x54, 0x4d, 0x28, 0x26, 0xd9, 0xaa, 0x60, 0x63,
	0x6b, 0x40, 0x37, 0x46, 0x26, 0x78, 0x0a, 0xbb,
})

// testNet3GenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the test network (version 3).  It is the same as the merkle root
// for the main network.
var testNet3GenesisMerkleRoot = genesisMerkleRoot

var testNet3GenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:    1,
		PrevBlock:  chainhash.Hash{},          // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: testNet3GenesisMerkleRoot, // 5b2a3f53f605d62c53e62932dac6925e3d74afa5a4b459745c36d42d0ed26a69
		Timestamp:  time.Unix(1391503289, 0),  // 2014-02-04 08:41:29 +0000 UTC
		Bits:       0x1e0ffff0,                // 504365040 [00000ffff0000000000000000000000000000000000000000000000000000000]
		Nonce:      0xf39f7,                   // 997879
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

var bigOne = big.NewInt(1)
var testNet3PowLimit = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 236), bigOne)

var DogeCoinTestNet3Params = chaincfg.Params{
	Name:        "testnet3",
	Net:         TestNet3,
	DefaultPort: "44556",
	DNSSeeds: []chaincfg.DNSSeed{
		{Host: "testseed.jrn.me.uk", HasFiltering: false},
	},

	GenesisBlock:             &testNet3GenesisBlock,
	GenesisHash:              &testNet3GenesisHash,
	PowLimit:                 testNet3PowLimit,
	PowLimitBits:             0x3bffffff,
	BIP0034Height:            708658,
	BIP0065Height:            1854705,
	BIP0066Height:            708658,
	CoinbaseMaturity:         240,
	SubsidyReductionInterval: 100000,
	TargetTimespan:           time.Minute,
	TargetTimePerBlock:       time.Minute,
	RetargetAdjustmentFactor: 4,
	ReduceMinDifficulty:      true,
	MinDiffReductionTime:     time.Minute * 2,
	GenerateSupported:        false,

	Checkpoints: []chaincfg.Checkpoint{
		{Height: 0, Hash: newHashFromStr("bb0a78264637406b6360aad926284d544d7049f45189db5664f3c4d07350559e")},
		{Height: 483173, Hash: newHashFromStr("a804201ca0aceb7e937ef7a3c613a9b7589245b10cc095148c4ce4965b0b73b5")},
		{Height: 591117, Hash: newHashFromStr("5f6b93b2c28cedf32467d900369b8be6700f0649388a7dbfd3ebd4a01b1ffad8")},
		{Height: 658924, Hash: newHashFromStr("ed6c8324d9a77195ee080f225a0fca6346495e08ded99bcda47a8eea5a8a620b")},
		{Height: 703635, Hash: newHashFromStr("839fa54617adcd582d53030a37455c14a87a806f6615aa8213f13e196230ff7f")},
		{Height: 1000000, Hash: newHashFromStr("1fe4d44ea4d1edb031f52f0d7c635db8190dc871a190654c41d2450086b8ef0e")},
		{Height: 1202214, Hash: newHashFromStr("a2179767a87ee4e95944703976fee63578ec04fa3ac2fc1c9c2c83587d096977")},
		{Height: 1250000, Hash: newHashFromStr("b46affb421872ca8efa30366b09694e2f9bf077f7258213be14adb05a9f41883")},
		{Height: 1500000, Hash: newHashFromStr("0caa041b47b4d18a4f44bdc05cef1a96d5196ce7b2e32ad3e4eb9ba505144917")},
		{Height: 1750000, Hash: newHashFromStr("8042462366d854ad39b8b95ed2ca12e89a526ceee5a90042d55ebb24d5aab7e9")},
		{Height: 2000000, Hash: newHashFromStr("d6acde73e1b42fc17f29dcc76f63946d378ae1bd4eafab44d801a25be784103c")},
		{Height: 2250000, Hash: newHashFromStr("c4342ae6d9a522a02e5607411df1b00e9329563ef844a758d762d601d42c86dc")},
		{Height: 2500000, Hash: newHashFromStr("3a66ec4933fbb348c9b1889aaf2f732fe429fd9a8f74fee6895eae061ac897e2")},
		{Height: 2750000, Hash: newHashFromStr("473ea9f625d59f534ffcc9738ffc58f7b7b1e0e993078614f5484a9505885563")},
		{Height: 3062910, Hash: newHashFromStr("113c41c00934f940a41f99d18b2ad9aefd183a4b7fe80527e1e6c12779bd0246")},
	},

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

	Bech32HRPSegwit: "tdge",

	PubKeyHashAddrID:        0x71,
	ScriptHashAddrID:        0xc4,
	PrivateKeyID:            0xf1,
	WitnessPubKeyHashAddrID: 0x00,
	WitnessScriptHashAddrID: 0x00,

	HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xcf},
	HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x94},

	HDCoinType: 1,
}
