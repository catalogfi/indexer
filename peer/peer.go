package peer

import (
	"fmt"
	"math"
	"net"
	"time"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/peer"
	"github.com/btcsuite/btcd/wire"
	"github.com/catalogfi/indexer/model"
	"gorm.io/gorm"
)

type Storage interface {
	GetBlockLocator() (blockchain.BlockLocator, error)
	PutBlock(block *wire.MsgBlock) error
	PutTx(tx *wire.MsgTx) error
	Params() *chaincfg.Params
}

type Peer struct {
	peer    *peer.Peer
	storage Storage
}

func NewPeer(str Storage) (*Peer, error) {
	done := make(chan struct{})

	peerCfg := &peer.Config{
		UserAgentName:    "peer",  // User agent name to advertise.
		UserAgentVersion: "1.0.0", // User agent version to advertise.
		ChainParams:      str.Params(),
		Services:         wire.SFNodeWitness,
		TrickleInterval:  time.Second * 10,
		Listeners: peer.MessageListeners{
			OnInv: func(p *peer.Peer, msg *wire.MsgInv) {
				sendMsg := wire.NewMsgGetData()
				for _, inv := range msg.InvList {
					sendMsg.AddInvVect(inv)
					fmt.Println("got an inv", inv.Type.String())
				}
				p.QueueMessage(sendMsg, done)
			},
			OnBlock: func(p *peer.Peer, msg *wire.MsgBlock, buf []byte) {

				if err := str.PutBlock(msg); err != nil {
					fmt.Printf("error putting block (%s): %v\n", msg.BlockHash().String(), err)
				}
			},
			OnTx: func(p *peer.Peer, tx *wire.MsgTx) {
				fmt.Println("got a tx")
				if err := str.PutTx(tx); err != nil {
					fmt.Printf("error putting tx (%s): %v\n", tx.TxHash().String(), err)
				}
			},
		},
		AllowSelfConns: true,
	}

	p, err := peer.NewOutboundPeer(peerCfg, "127.0.0.1:18444")
	if err != nil {
		return nil, fmt.Errorf("NewOutboundPeer: error %v", err)
	}

	// Establish the connection to the peer address and mark it connected.
	conn, err := net.Dial("tcp", p.Addr())
	if err != nil {
		return nil, fmt.Errorf("net.Dial: error %v", err)
	}
	p.AssociateConnection(conn)

	return &Peer{
		peer:    p,
		storage: str,
	}, nil
}

func (p *Peer) Run() error {
	for {
		locator, err := p.storage.GetBlockLocator()
		if err != nil {
			return fmt.Errorf("GetBlockLocator: error %v", err)
		}
		if err := p.peer.PushGetBlocksMsg(locator, &chainhash.Hash{}); err != nil {
			return fmt.Errorf("PushGetBlocksMsg: error %v", err)
		}

		time.Sleep(time.Minute * 10)
		// once blocks are recieved, the OnBlock callback will be called
	}
}

func getLatestBlockHeight(db *gorm.DB) (int32, error) {
	block := &model.Block{}
	if resp := db.Order("height desc").First(block); resp.Error != nil {
		if resp.Error == gorm.ErrRecordNotFound {
			return -1, nil
		}
		return -1, fmt.Errorf("db error: %v", resp.Error)
	}
	return block.Height, nil
}

func getBlockLocator(db *gorm.DB) (blockchain.BlockLocator, error) {
	height, err := getLatestBlockHeight(db)
	if err != nil {
		return nil, err
	}
	locatorIDs := calculateLocator([]int{int(height)})
	blocks := []model.Block{}

	if res := db.Find(&blocks, "height in ?", locatorIDs); res.Error != nil {
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
