package peer

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

func NewPeer(url string, str Storage) (*Peer, error) {
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

	p, err := peer.NewOutboundPeer(peerCfg, url)
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
