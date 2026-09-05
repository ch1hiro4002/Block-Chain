package network

import (
	"bytes"
	"fmt"
	"sync"
)

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

func (lt *LocalTransport) SendMessage(to NetAddr, mt MessageType, payload []byte) error {
	lt.lock.RLock()
	defer lt.lock.RUnlock()

	peer, ok := lt.peers[to]
	if !ok {
		return fmt.Errorf("peer %s not connected", to)
	}

	msg := NewMessage(mt, payload)
	msgData := msg.Bytes()

	peer.consumeCh <- RPC{
		From:    lt.addr,
		Payload: bytes.NewReader(msgData),
	}

	return nil
}

func (lt *LocalTransport) Addr() NetAddr {
	return lt.addr
}
