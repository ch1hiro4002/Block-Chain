package api

import (
	"encoding/hex"
	"fmt"

	"github.com/ch1hiro4002/Block-Chain/core"
	"github.com/ch1hiro4002/Block-Chain/crypto"
)

type SignatureResponse struct {
	R string `json:"r"`
	S string `json:"s"`
}

type TransactionResponse struct {
	Hash      string             `json:"hash"`
	Data      string             `json:"data"`
	From      string             `json:"from"`
	Signature *SignatureResponse `json:"signature"`
}

type BlockResponse struct {
	Hash          string                `json:"hash"`
	Version       uint32                `json:"version"`
	TxHash        string                `json:"txHash"`
	PrevBlockHash string                `json:"prevBlockHash"`
	Timestamp     int64                 `json:"timestamp"`
	Height        uint32                `json:"height"`
	Transactions  []TransactionResponse `json:"transactions"`
	Validator     string                `json:"validator"`
	Signature     *SignatureResponse    `json:"signature"`
}

func NewBlockResponse(block *core.Block) (*BlockResponse, error) {
	if block == nil {
		return nil, fmt.Errorf("block is nil")
	}
	if block.Header == nil {
		return nil, fmt.Errorf("block header is nil")
	}

	response := &BlockResponse{
		Hash:          block.Hash(core.BlockHasher{}).String(),
		Version:       block.Version,
		TxHash:        block.TxHash.String(),
		PrevBlockHash: block.PrevBlockHash.String(),
		Timestamp:     block.Timestamp,
		Height:        block.Height,
		Transactions:  make([]TransactionResponse, 0, len(block.Transactions)),
		Validator:     block.Validator.Address().String(),
	}

	if block.Signature != nil {
		response.Signature = newSignatureResponse(block.Signature)
	}

	for _, tx := range block.Transactions {
		if tx == nil {
			return nil, fmt.Errorf("block contains nil transaction")
		}

		txResponse, err := newTransactionResponse(tx)
		if err != nil {
			return nil, err
		}
		response.Transactions = append(response.Transactions, txResponse)
	}

	return response, nil
}

func newTransactionResponse(tx *core.Transaction) (TransactionResponse, error) {
	if tx == nil {
		return TransactionResponse{}, fmt.Errorf("transaction is nil")
	}

	response := TransactionResponse{
		Hash: tx.Hash(core.TxHasher{}).String(),
		Data: hex.EncodeToString(tx.Data),
		From: tx.From.Address().String(),
	}

	if tx.Signature != nil {
		response.Signature = newSignatureResponse(tx.Signature)
	}

	return response, nil
}

func newSignatureResponse(signature *crypto.Signature) *SignatureResponse {
	return &SignatureResponse{
		R: hex.EncodeToString(signature.R().Bytes()),
		S: hex.EncodeToString(signature.S().Bytes()),
	}
}
