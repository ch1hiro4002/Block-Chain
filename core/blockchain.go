package core

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

type Blockchain struct {
	store     Storage
	headers   []*Header
	validator Validator
	lock      sync.RWMutex
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
	bc.lock.Lock()
	defer bc.lock.Unlock()

	// Validate the block
	if err := bc.validator.ValidateBlock(block); err != nil {
		return err
	}

	return bc.addBlockWithoutValidation(block)
}

func (bc *Blockchain) Height() uint32 {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	return bc.heightLocked()
}

func (bc *Blockchain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

func (bc *Blockchain) GetHeader(height uint32) (*Header, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	return bc.getHeaderLocked(height)
}

func (bc *Blockchain) addBlockWithoutValidation(b *Block) error {
	bc.headers = append(bc.headers, b.Header)

	logrus.WithFields(logrus.Fields{
		"height":    b.Height,
		"hash":      b.Hash(BlockHasher{}),
		"prev_hash": b.PrevBlockHash,
	}).Info("Adding new block to blockchain")

	err := bc.store.Put(b)
	return err
}

func (bc *Blockchain) heightLocked() uint32 {
	if len(bc.headers) == 0 {
		return 0
	}

	return uint32(len(bc.headers) - 1)
}

func (bc *Blockchain) getHeaderLocked(height uint32) (*Header, error) {
	if height > bc.heightLocked() {
		return nil, fmt.Errorf("given height (%d) too high", height)
	}

	return bc.headers[height], nil
}
