package zcash

import (
	"fmt"
	"net"
	"time"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/peer"
	"github.com/btcsuite/btcd/wire"
)

type Storage interface {
	GetBlockLocator() (blockchain.BlockLocator, error)
	PutBlock(block *wire.MsgBlock) error
	PutTx(tx *wire.MsgTx) error
	Params() *chaincfg.Params
}

type Peer struct {
	done    chan struct{}
	peer    *peer.Peer
	storage Storage
}

type Network uint32

const (
	// Mainnet identifies the Zcash mainnet
	Mainnet Network = 0x6427e924
	// Testnet identifies ECC's public testnet
	Testnet Network = 0xbff91afa
	// Regtest identifies the regression test network
	Regtest Network = 0x5f3fe8aa
)

func NewPeer(url string, str Storage) (*Peer, error) {
	done := make(chan struct{})
	peerCfg := &peer.Config{
		UserAgentName:    "zfnd-seeder",
		UserAgentVersion: "0.1.3-alpha.4",
		ChainParams: &chaincfg.Params{
			Name:        "testnet",
			Net:         wire.BitcoinNet(Testnet),
			DefaultPort: "18233",
		},
		Services:        wire.SFNodeWitness,
		TrickleInterval: time.Second * 10,
		ProtocolVersion: 170100,
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
				// fmt.Println("got a block", p.LastAnnouncedBlock())
				fmt.Println("\n ### current block timeStamp: \n", time.Unix(msg.Header.Timestamp.Unix(), 0))
				if err := str.PutBlock(msg); err != nil {
					fmt.Printf("error putting block +++ (%s): %v\n", msg.BlockHash().String(), err)
				}
			},
			OnTx: func(p *peer.Peer, tx *wire.MsgTx) {
				fmt.Println("got a tx")
				fmt.Println("\n ### current tx timeStamp: \n", time.Unix(int64(tx.LockTime), 0))
				if err := str.PutTx(tx); err != nil {
					fmt.Printf("error putting tx (%s): %v\n", tx.TxHash().String(), err)
				}
			},
		},
		AllowSelfConns: true,
	}
	p, err := peer.NewOutboundPeer(peerCfg, url)
	if err != nil {
		return nil, fmt.Errorf("NewOutboundPeer: error %v", err)
	}

	// Establish the connection to the peer address and mark it connected.
	fmt.Println("url", p.Addr())
	conn, err := net.Dial("tcp", p.Addr())
	if err != nil {
		return nil, fmt.Errorf("net.Dial: error %v", err)
	}
	p.AssociateConnection(conn)

	return &Peer{
		done:    done,
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
		<-p.done
	}
}
