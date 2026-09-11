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
	TxHash        types.Hash
	PrevBlockHash types.Hash
	Timestamp     int64
	Height        uint32
}

type blockTransactionWire struct {
	Data      []byte
	From      crypto.PublicKey
	Signature *crypto.Signature
}

type blockWire struct {
	Header       blockHeaderWire
	Transactions []blockTransactionWire
	Validator    crypto.PublicKey
	Signature    *crypto.Signature
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
		Header: blockHeaderWire{
			Version:       block.Version,
			TxHash:        block.TxHash,
			PrevBlockHash: block.PrevBlockHash,
			Timestamp:     block.Timestamp,
			Height:        block.Height,
		},
		Transactions: make([]blockTransactionWire, 0, len(block.Transactions)),
	}

	wire.Validator = block.Validator

	if block.Signature != nil {
		wire.Signature = block.Signature
	}

	for _, tx := range block.Transactions {
		if tx == nil {
			return fmt.Errorf("cannot encode block with nil transaction")
		}

		wire.Transactions = append(wire.Transactions, blockTransactionWire{
			Data:      tx.Data,
			From:      tx.From,
			Signature: tx.Signature,
		})
	}

	return gob.NewEncoder(e.w).Encode(wire)
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

	block.Header = &Header{
		Version:       wire.Header.Version,
		TxHash:        wire.Header.TxHash,
		PrevBlockHash: wire.Header.PrevBlockHash,
		Timestamp:     wire.Header.Timestamp,
		Height:        wire.Header.Height,
	}
	block.Validator = wire.Validator

	block.Signature = wire.Signature

	block.Transactions = make([]*Transaction, 0, len(wire.Transactions))
	for _, txWire := range wire.Transactions {
		tx := &Transaction{
			Data: txWire.Data,
			From: txWire.From,
		}

		tx.Signature = txWire.Signature

		block.Transactions = append(block.Transactions, tx)
	}

	return nil
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
	w := struct {
		Data      []byte
		From      crypto.PublicKey
		Signature *crypto.Signature
	}{
		Data:      tx.Data,
		From:      tx.From,
		Signature: tx.Signature,
	}

	return gob.NewEncoder(e.w).Encode(w)
}

type Decoder[T any] interface {
	Decode(T) error
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
	w := struct {
		Data      []byte
		From      crypto.PublicKey
		Signature *crypto.Signature
	}{}

	if err := gob.NewDecoder(d.r).Decode(&w); err != nil {
		return err
	}

	tx.Data = w.Data
	tx.From = w.From

	tx.Signature = w.Signature

	return nil
}
