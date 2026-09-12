package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/ch1hiro4002/Block-Chain/types"
)

type Header struct {
	Version       uint32
	DataHash      types.Hash
	PrevBlockHash types.Hash
	Timestamp     int64
	Height        uint32
	Proposer      types.Address
}

type Block struct {
	*Header
	Transactions []*Transaction
	Validator    crypto.PublicKey
	Signature    *crypto.Signature
	hash         types.Hash
}

func NewBlock(h *Header, txs []*Transaction) (*Block, error) {
	return &Block{
		Header:       h,
		Transactions: txs,
	}, nil
}

func NewBlockFromPrevHeader(prevHeader *Header, txs []*Transaction) (*Block, error) {
	txHash, err := CalculateDataHash(txs)
	if err != nil {
		return nil, err
	}

	header := &Header{
		Version:       1,
		DataHash:      txHash,
		PrevBlockHash: BlockHasher{}.Hash(prevHeader),
		Timestamp:     time.Now().UnixNano(),
		Height:        prevHeader.Height + 1,
	}

	return NewBlock(header, txs)
}

func (b *Block) AddTransaction(tx *Transaction) {
	b.Transactions = append(b.Transactions, tx)
}

func (h *Header) Bytes() []byte {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	enc.Encode(headerToWire(h))

	return buf.Bytes()
}

func (b *Block) Sign(privateKey crypto.PrivateKey) error {
	sig, err := privateKey.Sign(b.Header.Bytes())
	if err != nil {
		return err
	}

	b.Signature = sig
	b.Validator = privateKey.PublicKey()

	return nil
}

func (b *Block) Verify(prevHeader *Header) error {
	if b.Signature == nil {
		return fmt.Errorf("block has no signature")
	}

	if !b.Signature.Verify(b.Validator, b.Header.Bytes()) {
		return fmt.Errorf("invalid signature")
	}

	if b.Header.Proposer != b.Validator.Address() {
		return fmt.Errorf("block proposer does not match block validator")
	}

	hash := BlockHasher{}.Hash(prevHeader)
	if hash != b.PrevBlockHash {
		return fmt.Errorf("block(%d) has invalid previous block hash", b.Height)
	}

	for _, tx := range b.Transactions {
		if err := tx.Verify(); err != nil {
			return err
		}
	}

	txHash, err := CalculateDataHash(b.Transactions)
	if err != nil {
		return fmt.Errorf("failed to calculate datahash: %v", err)
	}

	if txHash != b.DataHash {
		return fmt.Errorf("block (%s) has an invalid data hash", b.Hash(BlockHasher{}))
	}

	return nil
}

func (b *Block) Encode(enc Encoder[*Block]) error {
	return enc.Encode(b)
}

func (b *Block) Decode(dec Decoder[*Block]) error {
	return dec.Decode(b)
}

func (b *Block) Hash(hasher Hasher[*Header]) types.Hash {
	if b.hash.IsZero() {
		b.hash = hasher.Hash(b.Header)
	}

	return b.hash
}

func CalculateDataHash(txs []*Transaction) (types.Hash, error) {
	buf := &bytes.Buffer{}

	for _, tx := range txs {
		if _, err := buf.Write(tx.Hash(TxHasher{}).Bytes()); err != nil {
			return types.Hash{}, err
		}
	}

	return sha256.Sum256(buf.Bytes()), nil
}

func newGenesisBlock() (*Block, error) {
	header := &Header{
		Version:       1,
		DataHash:      types.Hash{},
		PrevBlockHash: types.Hash{},
		Timestamp:     0,
		Height:        0,
	}

	return NewBlock(header, []*Transaction{})
}
