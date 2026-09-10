package network

import (
	"bytes"
	"fmt"
	"sync"
)

type NetAddr string

type Transport interface {
	Consume() <-chan RPC
	Connect(Transport) error
	SendMessage(NetAddr, []byte) error
	Broadcast([]byte) error
	Has(NetAddr) bool
	Addr() NetAddr
}

type LocalTransport struct {
	addr      NetAddr                     // 节点网络地址
	consumeCh chan RPC                    // 接收节点RPC消息的channel
	lock      sync.RWMutex                // 读写锁
	peers     map[NetAddr]*LocalTransport // 已连接的节点列表
}

func NewLocalTransport(addr NetAddr) Transport {
	return &LocalTransport{
		addr:      addr,
		consumeCh: make(chan RPC, 100),
		peers:     make(map[NetAddr]*LocalTransport),
	}
}

func (lt *LocalTransport) Consume() <-chan RPC {
	return lt.consumeCh
}

func (lt *LocalTransport) Connect(tr Transport) error {
	lt.lock.Lock()
	defer lt.lock.Unlock()

	lt.peers[tr.Addr()] = tr.(*LocalTransport)
	return nil
}

func (lt *LocalTransport) SendMessage(to NetAddr, payload []byte) error {
	lt.lock.RLock()
	defer lt.lock.RUnlock()

	peer, ok := lt.peers[to]
	if !ok {
		return fmt.Errorf("%s: could not send message to not connect peer %s", lt.Addr(), to)
	}

	peer.consumeCh <- RPC{
		From:    lt.addr,
		Payload: bytes.NewReader(payload),
	}

	return nil
}

func (lt *LocalTransport) Broadcast(payload []byte) error {
	for _, peer := range lt.peers {
		if err := lt.SendMessage(peer.Addr(), payload); err != nil {
			return fmt.Errorf("failed to send message: %v", err)
		}
	}

	return nil
}

func (lt *LocalTransport) Has(addr NetAddr) bool {
	_, ok := lt.peers[addr]
	return ok
}

func (lt *LocalTransport) Addr() NetAddr {
	return lt.addr
}
