package network

import (
	"bytes"
	"os"
	"time"

	"github.com/ch1hiro4002/Block-Chain/core"
	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/go-kit/log"
)

var defaultBlockTime = 5 * time.Second

type ServerOpts struct {
	ID            string
	Logger        log.Logger
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	Transports    []Transport
	BlockTime     time.Duration
	PrivateKey    *crypto.PrivateKey
}

type Server struct {
	ServerOpts
	chain       *core.BlockChain
	memPool     *TxPool
	isValidator bool
	rpcCh       chan RPC
	quitCh      chan struct{}
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

	blockchain, err := core.NewBlockChain()
	if err != nil {
		return nil, err
	}
	server := &Server{
		ServerOpts:  opts,
		chain:       blockchain,
		memPool:     NewTxPool(),
		isValidator: opts.PrivateKey != nil,
		rpcCh:       make(chan RPC),
		quitCh:      make(chan struct{}, 1),
	}

	if server.RPCProcessor == nil {
		server.RPCProcessor = server
	}

	return server, nil
}

func (s *Server) Strat() {
	s.initTransports()

	if s.isValidator {
		go s.validatorLoop()
	}

free:
	for {
		select {
		case rpc := <-s.rpcCh:
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				s.Logger.Log("error", err)
			}

			if err := s.ProcessMessage(msg); err != nil {
				s.Logger.Log("error", err)
			}
		case <-s.quitCh:
			break free
		}
	}

	s.Logger.Log("msg", "Server shutdown!!!")
}

func (s *Server) initTransports() {
	for _, tr := range s.Transports {
		go func(tr Transport) {
			for rpc := range tr.Consume() {
				s.rpcCh <- rpc
			}
		}(tr)
	}
}

func (s *Server) validatorLoop() {
	ticker := time.NewTicker(s.BlockTime)

	s.Logger.Log(
		"msg", "Starting server validator loop",
		"blocktime", s.BlockTime,
	)

	

	for {
		<-ticker.C
		s.createNewBlock(nil)
	}
}

func (s *Server) ProcessMessage(msg *DecodeMessage) error {
	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.processTransaction(t)
	}

	return nil
}

func (s *Server) processTransaction(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})
	if s.memPool.HasTransaction(hash) {
		s.Logger.Log("msg", "transaction already in mempool")
		return nil
	}

	if err := tx.Verify(); err != nil {
		return err
	}

	tx.SetTime(time.Now())

	go s.broadcastTransaction(tx)

	s.Logger.Log(
		"msg", "adding new tx to mempool",
		"hash", hash,
		"mempoolLength", s.memPool.Len(),
	)

	return s.memPool.addTransaction(tx)
}

func (s *Server) broadcastTransaction(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeTx, buf.Bytes())
	return s.broadcast(msg.Bytes())
}

func (s *Server) broadcast(payload []byte) error {
	for _, ts := range s.Transports {
		if err := ts.Broadcast(payload); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) createNewBlock(txs []*core.Transaction) error {
	currentHeader, err := s.chain.GetHeader(s.chain.Height())
	if err != nil {
		return err
	}

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

	return nil
}
