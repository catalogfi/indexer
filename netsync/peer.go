package netsync

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/peer"
	"github.com/btcsuite/btcd/wire"
	"go.uber.org/zap"
)

type Peer struct {
	*peer.Peer
	chainParams *chaincfg.Params
	// this chan is used to signal that a block has been processed
	blockProcessed chan struct{}
	msg            chan interface{}
	logger         *zap.Logger
}

func NewPeer(url string, chainParams *chaincfg.Params, logger *zap.Logger) (*Peer, error) {
	done := make(chan struct{})
	peerMsg := make(chan interface{})
	peerCfg := &peer.Config{
		UserAgentName:    "peer",
		UserAgentVersion: "1.0.0",
		ChainParams:      chainParams,
		Services:         wire.SFNodeWitness,
		TrickleInterval:  time.Second * 10,
		Listeners: peer.MessageListeners{
			OnInv: func(p *peer.Peer, msg *wire.MsgInv) {
				sendMsg := wire.NewMsgGetData()
				for _, inv := range msg.InvList {
					if inv.Type == wire.InvTypeTx {
						logger.Info("received tx inv", zap.String("hash", inv.Hash.String()))
					}
					sendMsg.AddInvVect(inv)
				}
				p.QueueMessage(sendMsg, nil)
			},
			OnBlock: func(p *peer.Peer, msg *wire.MsgBlock, buf []byte) {
				peerMsg <- msg
			},
			OnTx: func(p *peer.Peer, tx *wire.MsgTx) {
				peerMsg <- tx
			},
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
		Peer:           p,
		blockProcessed: done,
		msg:            peerMsg,
		chainParams:    chainParams,
		logger:         logger,
	}, nil
}

func (p *Peer) OnMsg(ctx context.Context, handler func(msg interface{}) error) chan struct{} {
	closed := make(chan struct{})
	go func() {
		defer func() {
			closed <- struct{}{}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-p.msg:
				if !ok {
					return
				}
				if _, ok := msg.(*wire.MsgBlock); ok {
					p.blockProcessed <- struct{}{}
				}
				err := handler(msg)
				if err != nil {
					p.logger.Error("error handling block. Exiting", zap.Error(err))
					return
				}
			}
		}

	}()
	return closed
}

func (p *Peer) Reconnect() (*Peer, error) {
	p.Disconnect()
	return NewPeer(p.Addr(), p.chainParams, p.logger)
}
