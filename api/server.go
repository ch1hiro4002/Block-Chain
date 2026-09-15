package api

import (
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/ch1hiro4002/Block-Chain/core"
	"github.com/ch1hiro4002/Block-Chain/types"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/log"
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
	router := gin.Default()

	router.GET("/blocks/:blockIDorHash", s.HandleGetBlock)
	router.GET("/tx/:txHash", s.HandleGetTransaction)

	return router.Run(s.ListenAddr)
}

func (s *Server) HandleGetTransaction(ctx *gin.Context) {
	txHash := ctx.Param("txHash")

	if len(txHash) != 64 {
		logrus.Errorf("invalid transaction hash %q", txHash)
		ctx.JSON(http.StatusBadRequest, map[string]any{"error": "invalid transaction hash"})
		return
	}

	hashBytes, err := hex.DecodeString(txHash)
	if err != nil {
		logrus.Errorf("invalid transaction hash %q: %v", txHash, err)
		ctx.JSON(http.StatusBadRequest, map[string]any{"error": "invalid transaction hash"})
		return
	}

	tx := s.transactions.Get(types.HashFromBytes(hashBytes))
	if tx == nil {
		ctx.JSON(http.StatusNotFound, map[string]any{"error": "transaction not found"})
		return
	}

	response, err := newTransactionResponse(tx)
	if err != nil {
		logrus.Errorf("failed to serialize transaction %s: %v", txHash, err)
		ctx.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to serialize transaction"})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (s *Server) HandleGetBlock(ctx *gin.Context) {
	blockIDorHash := ctx.Param("blockIDorHash")

	if len(blockIDorHash) == 64 {
		hashBytes, err := hex.DecodeString(blockIDorHash)
		if err != nil {
			logrus.Errorf("invalid block hash %q: %v", blockIDorHash, err)
			ctx.JSON(http.StatusBadRequest, map[string]any{"error": "invalid block hash"})
			return
		}

		block, err := s.blockchain.GetBlockWithHash(types.HashFromBytes(hashBytes))
		if err != nil {
			logrus.Errorf("failed to get block with hash %s: %v", blockIDorHash, err)
			ctx.JSON(http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}

		s.writeBlockResponse(ctx, blockIDorHash, block)
		return
	}

	height, err := strconv.Atoi(blockIDorHash)
	if err != nil || height < 0 {
		logrus.Errorf("invalid block ID %q: %v", blockIDorHash, err)
		ctx.JSON(http.StatusBadRequest, map[string]any{"error": "invalid block ID"})
		return
	}

	block, err := s.blockchain.GetBlockWithHeight(uint32(height))
	if err != nil {
		logrus.Errorf("failed to get block at height %d: %v", height, err)
		ctx.JSON(http.StatusNotFound, map[string]any{"error": err.Error()})
		return
	}

	s.writeBlockResponse(ctx, blockIDorHash, block)
}

func (s *Server) writeBlockResponse(ctx *gin.Context, blockID string, block *core.Block) {
	response, err := NewBlockResponse(block)
	if err != nil {
		logrus.Errorf("failed to serialize block %s: %v", blockID, err)
		ctx.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to serialize block"})
		return
	}

	ctx.JSON(http.StatusOK, response)
}
