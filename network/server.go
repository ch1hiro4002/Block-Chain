package network

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/ch1hiro4002/Block-Chain/core"
	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/go-kit/log"
)

var defaultBlockTime = 2 * time.Second

type ServerOpts struct {
	ID            string
	Logger        log.Logger
	TCPTransport  *TCPTransport
	SeedNodes     []string
	ListenAddr    string
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	BlockTime     time.Duration
	PrivateKey    *crypto.PrivateKey
}

type Server struct {
	ServerOpts
	peerMu      sync.RWMutex
	peerMap     map[NetAddr]*TCPPeer
	chain       *core.BlockChain
	memPool     *TxPool
	isValidator bool
	rpcCh       chan RPC
	quitCh      chan struct{}
	peerCh      chan *TCPPeer
}

func NewServer(opts ServerOpts) (*Server, error) {
	if opts.BlockTime == time.Duration(0) {
		opts.BlockTime = defaultBlockTime
	}

	if opts.RPCDecodeFunc == nil {
		opts.RPCDecodeFunc = DefaultRPCDecoeFunc
	}

	if opts.Logger == nil {
		opts.Logger = log.NewLogfmtLogger(os.Stderr)
		opts.Logger = log.With(opts.Logger, "ID", opts.ID)
	}

	blockchain, err := core.NewBlockChain(opts.Logger)
	if err != nil {
		return nil, err
	}
	server := &Server{
		ServerOpts:  opts,
		peerMap:     make(map[NetAddr]*TCPPeer),
		chain:       blockchain,
		memPool:     NewTxPool(100),
		isValidator: opts.PrivateKey != nil,
		rpcCh:       make(chan RPC),
		quitCh:      make(chan struct{}, 1),
		peerCh:      opts.TCPTransport.peerCh,
	}

	if server.RPCProcessor == nil {
		server.RPCProcessor = server
	}

	// server.boostrapNodes()

	return server, nil
}

func (s *Server) Strat() {
	s.TCPTransport.Start()

	go s.bootstrapNetwork()

	if s.isValidator {
		go s.validatorLoop()
	}

free:
	for {
		select {
		case peer := <-s.peerCh:
			s.peerMu.Lock()
			s.peerMap[NetAddr(peer.conn.RemoteAddr().String())] = peer
			s.peerMu.Unlock()

			go peer.readLoop(s.rpcCh)

			if err := s.sendGetStatusMessage(peer); err != nil {
				s.Logger.Log("err", err)
				continue
			}

		case rpc := <-s.rpcCh:
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				s.Logger.Log("error", err)
				continue
			}

			if err := s.ProcessMessage(msg); err != nil {
				s.Logger.Log("error", err)
				continue
			}

		case <-s.quitCh:
			break free
		}
	}

	s.Logger.Log("msg", "Server shutdown!!!")
}

func (s *Server) bootstrapNetwork() {
	for _, nodeAddr := range s.SeedNodes {
		conn, err := net.Dial("tcp", nodeAddr)
		if err != nil {
			fmt.Printf("failed to connect: %v", err)
			continue
		}

		s.peerCh <- &TCPPeer{
			conn: conn,
		}
	}
}

func (s *Server) validatorLoop() {
	ticker := time.NewTicker(s.BlockTime)

	for {
		<-ticker.C
		if err := s.createNewBlock(); err != nil {
			s.Logger.Log("msg", "failed to create new block", "err", err)
		}
	}
}

func (s *Server) ProcessMessage(msg *DecodeMessage) error {
	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.processTransaction(t)
	case *core.Block:
		return s.processBlock(t)
	case *GetStatusMessage:
		return s.processGetStatusMessage(msg.From)
	case *StatusMessage:
		return s.processStatusMessage(msg.From, t)
	case *GetBlocksMessage:
		return s.processGetBlocksMessage(msg.From, t)
	case *BlocksMessage:
		return s.processBlocksMessage(msg.From, t)
	}

	return nil
}

func (s *Server) processTransaction(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})

	if s.memPool.Contains(hash) {
		// Silently handle duplicate transactions
		return nil
	}

	if err := tx.Verify(); err != nil {
		return err
	}

	tx.SetTime(time.Now())

	s.memPool.AddTransaction(tx)

	s.Logger.Log(
		"msg", "adding new tx to mempool",
		"hash", hash,
		"mempoolPengding", s.memPool.PendingCount(),
	)

	go s.broadcastTransaction(tx)

	return nil
}

func (s *Server) broadcastTransaction(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeTx, buf.Bytes())

	return s.broadcast(msg.Bytes())
}

func (s *Server) processBlock(block *core.Block) error {
	err := s.chain.AddBlock(block)
	if err != nil {
		if errors.Is(err, core.ErrBlockAlreadyExists) {
			return nil
		}

		return err
	}

	s.memPool.RemovePendingTransactions(block.Transactions)

	go s.broadcastBlock(block)

	return nil
}

func (s *Server) broadcastBlock(block *core.Block) error {
	buf := &bytes.Buffer{}
	err := block.Encode(core.NewGobBlockEncoder(buf))
	if err != nil {
		return err
	}

	msg := NewMessage(MessageTypeBlock, buf.Bytes())

	return s.broadcast(msg.Bytes())
}

func (s *Server) broadcast(payload []byte) error {
	s.peerMu.RLock()
	peers := make([]*TCPPeer, 0, len(s.peerMap))
	for _, peer := range s.peerMap {
		peers = append(peers, peer)
	}
	s.peerMu.RUnlock()

	for _, peer := range peers {
		if err := peer.Send(payload); err != nil {
			s.Logger.Log(
				"msg", "failed to send message",
				"to", peer.conn.RemoteAddr().String(),
				"err", err,
			)
		}
	}
	return nil
}

func (s *Server) sendGetStatusMessage(peer *TCPPeer) error {
	var (
		getStatusMessage = new(GetStatusMessage)
		buf              = new(bytes.Buffer)
	)

	if err := gob.NewEncoder(buf).Encode(getStatusMessage); err != nil {
		return nil
	}

	msg := NewMessage(MessageTypeGetStatus, buf.Bytes())

	return peer.Send(msg.Bytes())
}

func (s *Server) processGetStatusMessage(from NetAddr) error {
	s.Logger.Log(
		"msg", "received request status msg",
		"from", from,
	)

	statusMessage := &StatusMessage{
		ID:            s.ID,
		CurrentHeight: s.chain.Height(),
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(statusMessage); err != nil {
		return err
	}

	s.peerMu.RLock()
	peer, ok := s.peerMap[from]
	s.peerMu.RUnlock()
	if !ok {
		return fmt.Errorf("peer %s not known", from)
	}

	msg := NewMessage(MessageTypeStatus, buf.Bytes())

	return peer.Send(msg.Bytes())
}

func (s *Server) processStatusMessage(from NetAddr, data *StatusMessage) error {
	s.Logger.Log(
		"msg", "received STATUS message",
		"from", from,
	)

	if data.CurrentHeight <= s.chain.Height() {
		s.Logger.Log(
			"msg", "cannot to sync block to low",
			"height", s.chain.Height(),
			"other height", data.CurrentHeight,
		)
		return nil
	}

	getBlocksMessage := &GetBlocksMessage{
		From: s.chain.Height() + 1,
		To:   0,
	}

	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(getBlocksMessage)
	if err != nil {
		return err
	}

	s.peerMu.RLock()
	peer, ok := s.peerMap[from]
	s.peerMu.RUnlock()
	if !ok {
		return fmt.Errorf("peer %s not known", from)
	}

	msg := NewMessage(MessageTypeGetBlocks, buf.Bytes())

	return peer.Send(msg.Bytes())
}

func (s *Server) processGetBlocksMessage(from NetAddr, data *GetBlocksMessage) error {
	s.Logger.Log(
		"msg", "received get blocks message",
		"from", from,
	)

	blocks, err := s.chain.GetBlocks(data.From, data.To)
	if err != nil {
		return err
	}

	blocksMessage := &BlocksMessage{
		Blocks: blocks,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(blocksMessage); err != nil {
		return err
	}

	s.peerMu.RLock()
	peer, ok := s.peerMap[from]
	s.peerMu.RUnlock()
	if !ok {
		return fmt.Errorf("peer %s not known", from)
	}

	msg := NewMessage(MessageTypeBlocks, buf.Bytes())

	return peer.Send(msg.Bytes())
}

func (s *Server) processBlocksMessage(from NetAddr, data *BlocksMessage) error {
	s.Logger.Log(
		"msg", "received blocks message",
		"from", from,
	)

	for _, block := range data.Blocks {
		if err := s.chain.AddBlock(block); err != nil {
			if errors.Is(err, core.ErrBlockAlreadyExists) {
				continue
			}

			return err
		}
	}
	s.Logger.Log(
		"msg", "successfully sync blocks",
	)
	return nil
}

func (s *Server) createNewBlock() error {
	currentHeader, err := s.chain.GetHeader(s.chain.Height())
	if err != nil {
		return err
	}

	txs := s.memPool.Pending()

	block, err := core.NewBlockFromPrevHeader(currentHeader, txs)
	if err != nil {
		return err
	}

	if err := block.Sign(*s.PrivateKey); err != nil {
		return nil
	}

	if err := s.chain.AddBlock(block); err != nil {
		return err
	}

	s.memPool.ClearPending()

	s.broadcastBlock(block)

	return nil
}
