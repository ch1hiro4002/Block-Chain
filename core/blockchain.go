package core

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

type Blockchain struct {
	store     Storage
	headers   []*Header
	validator Validator
}

func NewBlockchain(genesis *Block) (*Blockchain, error) {
	bc := &Blockchain{
		store:   NewMemoryStore(),
		headers: []*Header{},
	}
	bc.validator = NewBlockValidator(bc)

	err := bc.addBlockWithoutValidation(genesis)

	return bc, err
}

func (bc *Blockchain) SetValidator(v Validator) {
	bc.validator = v
}

func (bc *Blockchain) AddBlock(block *Block) error {
	// Validate the block
	if err := bc.validator.ValidateBlock(block); err != nil {
		return err
	}

	return bc.addBlockWithoutValidation(block)
}

func (bc *Blockchain) Height() uint32 {
	return uint32(len(bc.headers) - 1)
}

func (bc *Blockchain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

func (bc *Blockchain) GetHeader(height uint32) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("given height (%d) too high", height)
	}

	return bc.headers[height], nil
}

func (bc *Blockchain) addBlockWithoutValidation(b *Block) error {
	bc.headers = append(bc.headers, b.Header)

	logrus.WithFields(logrus.Fields{
		"height":    b.Height,
		"hash":      b.Hash(BlockHasher{}),
		"prev_hash": b.PrevBlockHash,
	}).Info("Adding new block to blockchain")

	return bc.store.Put(b)
}
