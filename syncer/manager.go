package syncer

import (
	"github.com/catalogfi/indexer/peer"
	"go.uber.org/zap"
)

type SyncManager struct {
	peer   *peer.Peer //TODO: will we have multiple peers in future?
	logger *zap.Logger
}

func NewSyncManager(peer *peer.Peer, logger *zap.Logger) *SyncManager {
	return &SyncManager{peer: peer, logger: logger}
}

func (s *SyncManager) Sync() {
	for {
		go s.peer.Run()
		s.peer.WaitForDisconnect()
		s.logger.Info("Disconnected from peer, reconnecting...", zap.String("peer", s.peer.Addr()))

		reconnectedPeer, err := s.peer.Reconnect()
		if err != nil {
			//TODO: need to recover from this if this happens too often
			s.logger.Error("error reconnecting to peer", zap.String("peer", s.peer.Addr()), zap.Error(err))
			continue
		}
		s.peer = reconnectedPeer
	}
}
