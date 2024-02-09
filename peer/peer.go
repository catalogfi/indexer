package peer

// import (
// 	"fmt"

// 	"github.com/btcsuite/btcd/chaincfg"
// 	"github.com/btcsuite/btcd/chaincfg/chainhash"
// 	"github.com/btcsuite/btcd/peer"
// 	"github.com/btcsuite/btcd/wire"
// 	"github.com/catalogfi/indexer/database"
// )

// type Peer struct {
// 	done chan struct{}
// 	peer *peer.Peer
// 	db   database.Db
// }

// func NewPeer(url string, chainParams *chaincfg.Params, db database.Db) (*Peer, error) {

// 	return &Peer{
// 		done: done,
// 		peer: p,
// 		db:   db,
// 	}, nil
// }

// func (p *Peer) WaitForDisconnect() {
// 	p.peer.WaitForDisconnect()
// }

// func (p *Peer) Addr() string {
// 	return p.peer.Addr()
// }

// func (p *Peer) Reconnect() (*Peer, error) {
// 	p.peer.Disconnect()

// 	peerAddr := p.peer.Addr()
// 	storage := p.storage

// 	peer, err := NewPeer(peerAddr, storage)
// 	if err != nil {
// 		return nil, fmt.Errorf("error reconnecting to peer: %v\n", err)
// 	}
// 	return peer, nil
// }

// func (p *Peer) Run() error {

// 	for {
// 		if !p.peer.Connected() {
// 			close(p.done)
// 			return fmt.Errorf("peer disconnected")
// 		}
// 		locator, err := p.storage.GetBlockLocator()
// 		if err != nil {
// 			return fmt.Errorf("GetBlockLocator: error %v", err)
// 		}
// 		if len(locator) > 0 {
// 			fmt.Println("sending getblocks", locator[0].String())
// 		}
// 		if err := p.peer.PushGetBlocksMsg(locator, &chainhash.Hash{}); err != nil {
// 			return fmt.Errorf("PushGetBlocksMsg: error %v", err)
// 		}
// 		<-p.done
// 	}
// }

// func (p *Peer) putBlock(block *wire.MsgBlock) error {
// 	return p.storage.PutBlock(block)
// }
