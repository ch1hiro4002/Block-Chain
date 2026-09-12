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

	"github.com/ch1hiro4002/Block-Chain/api"
	"github.com/ch1hiro4002/Block-Chain/core"
	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/ch1hiro4002/Block-Chain/types"
	"github.com/go-kit/log"
)

var defaultBlockTime = 2 * time.Second

type ServerOpts struct {
	ID            string
	Logger        log.Logger
	TCPTransport  *TCPTransport
	SeedNodes     []string
	ListenAddr    string
	APIListenAddr string
	RPCDecodeFunc RPCDecodeFunc
	BlockTime     time.Duration
	PrivateKey    *crypto.PrivateKey
}

type Server struct {
	ServerOpts
	peerMu      sync.RWMutex
	peerMap     map[NetAddr]*TCPPeer
	syncMu      sync.RWMutex
	syncTarget  uint32
	synced      bool
	chain       *core.BlockChain
	memPool     *TxPool
	isValidator bool
	rpcCh       chan RPC
	quitCh      chan struct{}
	peerCh      chan *TCPPeer
}

func NewServer(opts ServerOpts, addr types.Address, pubKey crypto.PublicKey, stake uint64) (*Server, error) {
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

	blockchain, err := core.NewBlockChain(opts.Logger, addr, pubKey, stake)
	if err != nil {
		return nil, err
	}

	memPool := NewTxPool(100)

	// Config API services
	if len(opts.APIListenAddr) > 0 {
		apiServerConfig := api.ServerConfig{
			Logger:     opts.Logger,
			ListenAddr: opts.APIListenAddr,
		}

		apiServer := api.NewServer(apiServerConfig, blockchain, memPool)
		go apiServer.Start()

		apiServer.Logger.Log(
			"msg", "JSON API server running",
			"port", apiServer.ListenAddr,
		)
	}

	server := &Server{
		ServerOpts:  opts,
		peerMap:     make(map[NetAddr]*TCPPeer),
		chain:       blockchain,
		memPool:     memPool,
		isValidator: opts.PrivateKey != nil,
		synced:      len(opts.SeedNodes) == 0,
		rpcCh:       make(chan RPC),
		quitCh:      make(chan struct{}, 1),
		peerCh:      opts.TCPTransport.peerCh,
	}

	return server, nil
}

func (s *Server) isSynced() bool {
	s.syncMu.RLock()
	defer s.syncMu.RUnlock()

	return s.synced
}

func (s *Server) markSynced() {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()

	s.synced = true
	s.syncTarget = 0
}

func (s *Server) setSyncTarget(height uint32) {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()

	s.synced = false
	if height > s.syncTarget {
		s.syncTarget = height
	}
}

func (s *Server) markSyncProgressDone() {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()

	if s.syncTarget == 0 || s.chain.Height() >= s.syncTarget {
		s.synced = true
		s.syncTarget = 0
	}
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

			if err := s.sendGetValidatorsMessage(peer); err != nil {
				s.Logger.Log("err", err)
				continue
			}

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

		if !s.isSynced() {
			continue
		}

		if !s.shouldPropose() {
			continue
		}

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
	case *GetValidatorsMessage:
		return s.processGetValidatorsMessage(msg.From)
	case *ValidatorsMessage:
		return s.processValidatorsMessage(msg.From, t)
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

	if err := s.chain.ValidateTransaction(tx, time.Now()); err != nil {
		return err
	}

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

func (s *Server) sendGetValidatorsMessage(peer *TCPPeer) error {
	getValidatorsMessage := new(GetValidatorsMessage)
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(getValidatorsMessage); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeGetValidators, buf.Bytes())

	return peer.Send(msg.Bytes())
}

func (s *Server) processGetValidatorsMessage(from NetAddr) error {
	validators := s.chain.GetValidatorSet().Snapshot()
	validatorsMessage := &ValidatorsMessage{
		Height:     s.chain.Height(),
		Validators: validators,
		History:    s.chain.GetValidatorSetHistory(),
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(validatorsMessage); err != nil {
		return err
	}

	s.peerMu.RLock()
	peer, ok := s.peerMap[from]
	s.peerMu.RUnlock()
	if !ok {
		return fmt.Errorf("peer %s not known", from)
	}

	msg := NewMessage(MessageTypeValidators, buf.Bytes())

	return peer.Send(msg.Bytes())
}

func (s *Server) processValidatorsMessage(from NetAddr, data *ValidatorsMessage) error {
	if data == nil {
		return fmt.Errorf("received nil validators message")
	}

	if err := s.chain.SetValidatorSetHistory(data.History); err != nil {
		return fmt.Errorf("failed to import validator history from %s: %w", from, err)
	}

	if err := s.chain.GetValidatorSet().Merge(data.Validators); err != nil {
		return fmt.Errorf("failed to merge validators from %s: %w", from, err)
	}

	s.Logger.Log(
		"msg", "validators updated",
		"from", from,
		"height", data.Height,
		"validators", len(data.Validators),
	)

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
		s.markSynced()
		return nil
	}

	s.setSyncTarget(data.CurrentHeight)

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
	s.markSyncProgressDone()
	return nil
}

func (s *Server) createNewBlock() error {
	currentHeader, err := s.chain.GetHeader(s.chain.Height())
	if err != nil {
		return err
	}

	txs := s.memPool.Pending()

	proposer := s.PrivateKey.PublicKey()

	block, err := core.NewBlockFromPrevHeader(currentHeader, txs)
	if err != nil {
		return err
	}

	block.Header.Proposer = proposer.Address()

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

func (s *Server) shouldPropose() bool {
	if s.PrivateKey == nil {
		return false
	}

	nextHeight := s.chain.Height() + 1
	currentHeader, err := s.chain.GetHeader(s.chain.Height())
	if err != nil {
		return false
	}

	prevHash := core.BlockHasher{}.Hash(currentHeader)
	proposer := s.chain.GetValidatorSet().Proposer(nextHeight, prevHash)

	return proposer == s.PrivateKey.PublicKey().Address()
}
