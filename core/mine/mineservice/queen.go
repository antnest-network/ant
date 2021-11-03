package mineservice

import (
	"context"
	"errors"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	ant_pro "github.com/antnest-network/ant-proto/pb"
	"math/rand"
	"sync"
)

type QueenManager struct {
	bootstrapQueens []peer.AddrInfo
	activeQueens    []peer.AddrInfo
	currentIndex    int
	sync.RWMutex
}

func NewQueenManager(bootstrapQueens []peer.AddrInfo) *QueenManager {
	m := QueenManager{
		bootstrapQueens: bootstrapQueens,
	}
	return &m
}

func (m *QueenManager) HandleQueenMessage(ctx context.Context, from peer.ID, msg interface{}) {
	req, ok := msg.(*ant_pro.Queens)
	if !ok {
		log.Infof("msg type error: %v, %+v", from, msg)
		return
	}
	log.Infof("received queen message from %v", from)
	var queens []peer.AddrInfo
	for _, v := range req.Queens {
		addr, err := pbQueen2Address(v)
		if err != nil {
			continue
		}
		queens = append(queens, addr)
	}
	m.Lock()
	m.activeQueens = queens
	m.Unlock()
}

func pbQueen2Address(p *ant_pro.Queens_Queen) (peer.AddrInfo, error) {
	ret := peer.AddrInfo{}
	for _, v := range p.Addrs {
		addr, err := ma.NewMultiaddr(v)
		if err != nil {
			log.Errorf("NewMultiaddr: %v", err)
			continue
		}
		ret.Addrs = append(ret.Addrs, addr)
	}
	if len(ret.Addrs) == 0 {
		return ret, errors.New("no valid address")
	}
	ret.ID = peer.ID(p.Id)
	return ret, nil
}

func (m *QueenManager) GetQueen() peer.AddrInfo {
	m.RLock()
	defer m.RUnlock()

	if len(m.activeQueens) > 0 {
		m.currentIndex = (m.currentIndex + 1) % len(m.activeQueens)
		return m.activeQueens[m.currentIndex]
	}
	return m.bootstrapQueens[rand.Intn(len(m.bootstrapQueens))]
}
