package api

import (
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/ch1hiro4002/Block-Chain/core"
	"github.com/ch1hiro4002/Block-Chain/types"
	"github.com/go-kit/log"
	"github.com/labstack/echo/v5"
	"github.com/sirupsen/logrus"
)

type ServerConfig struct {
	Logger     log.Logger
	ListenAddr string
}

type TransactionGetter interface {
	Get(types.Hash) *core.Transaction
}

type Server struct {
	ServerConfig
	blockchain   *core.BlockChain
	transactions TransactionGetter
}

func NewServer(config ServerConfig, blockchain *core.BlockChain, transactions TransactionGetter) *Server {
	return &Server{
		ServerConfig: config,
		blockchain:   blockchain,
		transactions: transactions,
	}
}

func (s *Server) Start() error {
	e := echo.New()

	e.GET("/blocks/:blockIDorHash", s.HandleGetBlock)
	e.GET("/tx/:txHash", s.HandleGetTransaction)

	return e.Start(s.ListenAddr)
}

func (s *Server) HandleGetTransaction(context *echo.Context) error {
	txHash := context.Param("txHash")

	if len(txHash) != 64 {
		logrus.Errorf("invalid transaction hash %q", txHash)
		return context.JSON(http.StatusBadRequest, map[string]any{"error": "invalid transaction hash"})
	}

	hashBytes, err := hex.DecodeString(txHash)
	if err != nil {
		logrus.Errorf("invalid transaction hash %q: %v", txHash, err)
		return context.JSON(http.StatusBadRequest, map[string]any{"error": "invalid transaction hash"})
	}

	tx := s.transactions.Get(types.HashFromBytes(hashBytes))
	if tx == nil {
		return context.JSON(http.StatusNotFound, map[string]any{"error": "transaction not found"})
	}

	response, err := newTransactionResponse(tx)
	if err != nil {
		logrus.Errorf("failed to serialize transaction %s: %v", txHash, err)
		return context.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to serialize transaction"})
	}

	return context.JSON(http.StatusOK, response)
}

func (s *Server) HandleGetBlock(context *echo.Context) error {
	blockIDorHash := context.Param("blockIDorHash")

	if len(blockIDorHash) == 64 {
		hashBytes, err := hex.DecodeString(blockIDorHash)
		if err == nil {
			block, err := s.blockchain.GetBlockWithHash(types.HashFromBytes(hashBytes))
			if err != nil {
				logrus.Errorf("failed to get block with hash %s: %v", blockIDorHash, err)
				return context.JSON(http.StatusNotFound, map[string]any{"error": err.Error()})
			}

			return s.writeBlockResponse(context, blockIDorHash, block)
		}
	}

	height, err := strconv.Atoi(blockIDorHash)
	if err != nil || height < 0 {
		logrus.Errorf("invalid block ID %q: %v", blockIDorHash, err)
		return context.JSON(http.StatusBadRequest, map[string]any{"error": "invalid block ID"})
	}

	block, err := s.blockchain.GetBlockWithHeight(uint32(height))
	if err != nil {
		logrus.Errorf("failed to get block at height %d: %v", height, err)
		return context.JSON(http.StatusNotFound, map[string]any{"error": err.Error()})
	}

	return s.writeBlockResponse(context, blockIDorHash, block)
}

func (s *Server) writeBlockResponse(context *echo.Context, blockID string, block *core.Block) error {
	response, err := NewBlockResponse(block)
	if err != nil {
		logrus.Errorf("failed to serialize block %s: %v", blockID, err)
		return context.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to serialize block"})
	}

	return context.JSON(http.StatusOK, response)
}
