package core

import (
	"encoding/gob"
	"fmt"
	"io"

	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/ch1hiro4002/Block-Chain/types"
)

type Encoder[T any] interface {
	Encode(T) error
}

type blockHeaderWire struct {
	Version       uint32
	DataHash      types.Hash
	PrevBlockHash types.Hash
	Timestamp     int64
	Height        uint32
}

type transactionWire struct {
	Nonce     uint64
	Timestamp int64
	Deadline  int64
	Data      []byte
	From      crypto.PublicKey
	Signature *crypto.Signature
}

type blockWire struct {
	Header       blockHeaderWire
	Transactions []transactionWire
	Validator    crypto.PublicKey
	Signature    *crypto.Signature
}

func headerToWire(h *Header) blockHeaderWire {
	return blockHeaderWire{
		Version:       h.Version,
		DataHash:      h.DataHash,
		PrevBlockHash: h.PrevBlockHash,
		Timestamp:     h.Timestamp,
		Height:        h.Height,
	}
}

func headerFromWire(w blockHeaderWire) *Header {
	return &Header{
		Version:       w.Version,
		DataHash:      w.DataHash,
		PrevBlockHash: w.PrevBlockHash,
		Timestamp:     w.Timestamp,
		Height:        w.Height,
	}
}

func transactionToWire(tx *Transaction) transactionWire {
	return transactionWire{
		Nonce:     tx.Nonce,
		Timestamp: tx.Timestamp,
		Deadline:  tx.Deadline,
		Data:      tx.Data,
		From:      tx.From,
		Signature: tx.Signature,
	}
}

func transactionFromWire(w transactionWire) *Transaction {
	return &Transaction{
		Nonce:     w.Nonce,
		Timestamp: w.Timestamp,
		Deadline:  w.Deadline,
		Data:      w.Data,
		From:      w.From,
		Signature: w.Signature,
	}
}

type GobBlockEncoder struct {
	w io.Writer
}

func NewGobBlockEncoder(w io.Writer) *GobBlockEncoder {
	return &GobBlockEncoder{
		w: w,
	}
}

func (e *GobBlockEncoder) Encode(block *Block) error {
	if block == nil {
		return fmt.Errorf("cannot encode nil block")
	}
	if block.Header == nil {
		return fmt.Errorf("cannot encode block with nil header")
	}

	wire := blockWire{
		Header:       headerToWire(block.Header),
		Transactions: make([]transactionWire, 0, len(block.Transactions)),
	}

	wire.Validator = block.Validator

	if block.Signature != nil {
		wire.Signature = block.Signature
	}

	for _, tx := range block.Transactions {
		if tx == nil {
			return fmt.Errorf("cannot encode block with nil transaction")
		}

		wire.Transactions = append(wire.Transactions, transactionToWire(tx))
	}

	return gob.NewEncoder(e.w).Encode(wire)
}

type GobTxEncoder struct {
	w io.Writer
}

func NewGobTxEncoder(w io.Writer) *GobTxEncoder {
	return &GobTxEncoder{
		w: w,
	}
}

func (e *GobTxEncoder) Encode(tx *Transaction) error {
	if tx == nil {
		return fmt.Errorf("cannot encode nil transaction")
	}

	return gob.NewEncoder(e.w).Encode(transactionToWire(tx))
}

type Decoder[T any] interface {
	Decode(T) error
}

type GobBlockDecoder struct {
	r io.Reader
}

func NewGobBlockDecoder(r io.Reader) *GobBlockDecoder {
	return &GobBlockDecoder{
		r: r,
	}
}

func (e *GobBlockDecoder) Decode(block *Block) error {
	if block == nil {
		return fmt.Errorf("cannot decode into nil block")
	}

	var wire blockWire
	if err := gob.NewDecoder(e.r).Decode(&wire); err != nil {
		return err
	}

	block.Header = headerFromWire(wire.Header)
	block.Validator = wire.Validator

	block.Signature = wire.Signature
	block.hash = types.Hash{}

	block.Transactions = make([]*Transaction, 0, len(wire.Transactions))
	for _, txWire := range wire.Transactions {
		block.Transactions = append(block.Transactions, transactionFromWire(txWire))
	}

	return nil
}

type GobTxDecoder struct {
	r io.Reader
}

func NewGobTxDecoder(r io.Reader) *GobTxDecoder {
	return &GobTxDecoder{
		r: r,
	}
}

func (d *GobTxDecoder) Decode(tx *Transaction) error {
	if tx == nil {
		return fmt.Errorf("cannot decode into nil transaction")
	}

	var w transactionWire
	if err := gob.NewDecoder(d.r).Decode(&w); err != nil {
		return err
	}

	*tx = *transactionFromWire(w)

	return nil
}
