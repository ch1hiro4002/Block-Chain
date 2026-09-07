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
	From      []byte
	Signature []byte
}

type blockWire struct {
	Header       blockHeaderWire
	Transactions []blockTransactionWire
	Validator    []byte
	Signature    []byte
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

	validator, err := block.Validator.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal block validator: %w", err)
	}
	wire.Validator = validator

	if block.Signature != nil {
		wire.Signature, err = block.Signature.Marshal()
		if err != nil {
			return fmt.Errorf("failed to marshal block signature: %w", err)
		}
	}

	for _, tx := range block.Transactions {
		if tx == nil {
			return fmt.Errorf("cannot encode block with nil transaction")
		}

		from, err := tx.From.Marshal()
		if err != nil {
			return fmt.Errorf("failed to marshal transaction sender: %w", err)
		}

		var sig []byte
		if tx.Signature != nil {
			sig, err = tx.Signature.Marshal()
			if err != nil {
				return fmt.Errorf("failed to marshal transaction signature: %w", err)
			}
		}

		wire.Transactions = append(wire.Transactions, blockTransactionWire{
			Data:      tx.Data,
			From:      from,
			Signature: sig,
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

	validator, err := crypto.UnmarshalPublicKey(wire.Validator)
	if err != nil {
		return fmt.Errorf("failed to unmarshal block validator: %w", err)
	}

	block.Header = &Header{
		Version:       wire.Header.Version,
		TxHash:        wire.Header.TxHash,
		PrevBlockHash: wire.Header.PrevBlockHash,
		Timestamp:     wire.Header.Timestamp,
		Height:        wire.Header.Height,
	}
	block.Validator = validator

	if len(wire.Signature) > 0 {
		sig, err := crypto.UnmarshalSignature(wire.Signature)
		if err != nil {
			return fmt.Errorf("failed to unmarshal block signature: %w", err)
		}
		block.Signature = &sig
	} else {
		block.Signature = nil
	}

	block.Transactions = make([]*Transaction, 0, len(wire.Transactions))
	for _, txWire := range wire.Transactions {
		from, err := crypto.UnmarshalPublicKey(txWire.From)
		if err != nil {
			return fmt.Errorf("failed to unmarshal transaction sender: %w", err)
		}

		tx := &Transaction{
			Data: txWire.Data,
			From: from,
		}

		if len(txWire.Signature) > 0 {
			sig, err := crypto.UnmarshalSignature(txWire.Signature)
			if err != nil {
				return fmt.Errorf("failed to unmarshal transaction signature: %w", err)
			}
			tx.Signature = &sig
		}

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
	from, err := tx.From.Marshal()
	if err != nil {
		return err
	}

	var sig []byte
	if tx.Signature != nil {
		sig, err = tx.Signature.Marshal()
		if err != nil {
			return err
		}
	}

	w := struct {
		Data      []byte
		From      []byte
		Signature []byte
	}{
		Data:      tx.Data,
		From:      from,
		Signature: sig,
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
		From      []byte
		Signature []byte
	}{}

	if err := gob.NewDecoder(d.r).Decode(&w); err != nil {
		return err
	}

	from, err := crypto.UnmarshalPublicKey(w.From)
	if err != nil {
		return err
	}

	tx.Data = w.Data
	tx.From = from

	if len(w.Signature) > 0 {
		sig, err := crypto.UnmarshalSignature(w.Signature)
		if err != nil {
			return err
		}
		tx.Signature = &sig
	}

	return nil
}
