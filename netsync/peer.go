package netsync

import (
	"fmt"
	"net"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/peer"
	"github.com/btcsuite/btcd/wire"
	"github.com/catalogfi/indexer/store"
	"go.uber.org/zap"
)

type Peer struct {
	*peer.Peer
	done        chan struct{}
	blocks      chan *wire.MsgBlock
	chainParams *chaincfg.Params
	logger      *zap.Logger
}

func NewPeer(url string, chainParams *chaincfg.Params, logger *zap.Logger) (*Peer, error) {
	done := make(chan struct{})
	blocks := make(chan *wire.MsgBlock)
	peerCfg := &peer.Config{
		UserAgentName:    "peer",
		UserAgentVersion: "1.0.0",
		ChainParams:      chainParams,
		Services:         wire.SFNodeWitness,
		TrickleInterval:  time.Second * 10,
		Listeners: peer.MessageListeners{

			//whenever we receive an inv message, we will request the data from the peer
			//inventory message is received when a peer requests (getblock) from another peer
			//and also when peer sends new mempool transactions
			OnInv: func(p *peer.Peer, msg *wire.MsgInv) {
				sendMsg := wire.NewMsgGetData()
				blockMsg := 0
				for _, inv := range msg.InvList {
					//TODO: handle tx invs
					if inv.Type == wire.InvTypeBlock {
						sendMsg.AddInvVect(inv)
						blockMsg++
					}
				}
				if blockMsg > 0 {
					p.QueueMessage(sendMsg, done)
				}
			},

			//whenever we receive a block message, we will put the block in our database
			OnBlock: func(p *peer.Peer, msg *wire.MsgBlock, buf []byte) {
				logger.Info("received block", zap.String("hash", msg.BlockHash().String()))
				blocks <- msg
			},
			//whenever we receive a tx message, we will put the tx in our database
			//this could get ignored if the blockchain is already syncing
			// OnTx: func(p *peer.Peer, tx *wire.MsgTx) {
			// 	if err := putTx(tx, config.Db); err != nil {
			// 		logger.Error("error putting tx", zap.Error(err))
			// 	}
			// },
		},
		AllowSelfConns: true,
	}
	p, err := peer.NewOutboundPeer(peerCfg, url)
	if err != nil {
		return nil, fmt.Errorf("syncManager: %v", err)
	}

	conn, err := net.Dial("tcp", p.Addr())
	if err != nil {
		return nil, fmt.Errorf("syncManager: %v", err)
	}

	p.AssociateConnection(conn)
	return &Peer{
		Peer:        p,
		done:        done,
		blocks:      blocks,
		chainParams: chainParams,
		logger:      logger,
	}, nil
}

func (p *Peer) OnBlock(handler func(block *wire.MsgBlock) error) {

	for {
		select {
		case block := <-p.blocks:
			err := handler(block)
			if err != nil && err.Error() != store.ErrKeyNotFound {
				p.logger.Error("error handling block. Exiting", zap.Error(err))
				return
			}
		}
	}

}

func (p *Peer) Reconnect() (*Peer, error) {
	p.Disconnect()
	return NewPeer(p.Addr(), p.chainParams, p.logger)
}
