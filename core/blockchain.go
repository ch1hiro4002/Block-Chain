package core

import (
	"fmt"
	"sync"

	"github.com/ch1hiro4002/Block-Chain/types"
	"github.com/go-kit/log"
)

type BlockChain struct {
	logger    log.Logger
	store     Storage
	headers   []*Header
	hashes    map[types.Hash]uint32
	validator Validator
	lock      sync.RWMutex
	state     *State
}

func NewBlockChain(logger log.Logger) (*BlockChain, error) {
	bc := &BlockChain{
		logger:  logger,
		store:   NewMemoryStore(),
		headers: []*Header{},
		hashes:  make(map[types.Hash]uint32),
		state:   NewState(),
	}
	bc.validator = NewBlockValidator(bc)

	genesis, err := newGenesisBlock()
	if err != nil {
		return nil, err
	}
	if err := bc.addBlockWithoutValidation(genesis); err != nil {
		return nil, err
	}

	return bc, nil
}

func (bc *BlockChain) AddBlock(block *Block) error {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	// Validate the block
	if err := bc.validator.ValidateBlock(block); err != nil {
		return err
	}

	for _, tx := range block.Transactions {
		bc.logger.Log(
			"msg", "executing code",
			"code length", len(tx.Data),
			"tx hash", tx.Hash(TxHasher{}),
		)

		vm := NewVM(tx.Data, bc.state)
		if err := vm.Run(); err != nil {
			return err
		}

		fmt.Printf("contract_state: %+v\n", vm.state)
	}

	return bc.addBlockWithoutValidation(block)
}

func (bc *BlockChain) Height() uint32 {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	return bc.heightLocked()
}

func (bc *BlockChain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

func (bc *BlockChain) GetHeader(height uint32) (*Header, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	return bc.getHeaderLocked(height)
}

func (bc *BlockChain) GetBlockWithHeight(height uint32) (*Block, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	if height > bc.heightLocked() {
		return nil, fmt.Errorf("given height (%d) too high", height)
	}

	return bc.store.Get(height)
}

func (bc *BlockChain) GetBlockWithHash(hash types.Hash) (*Block, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	height, ok := bc.hashes[hash]
	if !ok {
		return nil, fmt.Errorf("block with hash (%s) not found", hash)
	}

	return bc.store.Get(height)
}

func (bc *BlockChain) GetBlocks(from, to uint32) ([]*Block, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	if from > to && to != 0 {
		return nil, fmt.Errorf("invalid block range: from (%d) > to (%d)", from, to)
	}

	currentHeight := bc.heightLocked()

	if from > currentHeight {
		return []*Block{}, nil
	}

	if to > currentHeight || to == 0 {
		to = currentHeight
	}

	return bc.store.GetRange(from, to)
}

func (bc *BlockChain) addBlockWithoutValidation(b *Block) error {
	bc.headers = append(bc.headers, b.Header)
	hash := b.Hash(BlockHasher{})
	bc.hashes[hash] = b.Height

	bc.logger.Log(
		"msg", "new block",
		"hash", hash,
		"height", b.Height,
		"transactions", len(b.Transactions),
	)

	err := bc.store.Put(b)
	return err
}

func (bc *BlockChain) heightLocked() uint32 {
	if len(bc.headers) == 0 {
		return 0
	}

	return uint32(len(bc.headers) - 1)
}

func (bc *BlockChain) getHeaderLocked(height uint32) (*Header, error) {
	if height > bc.heightLocked() {
		return nil, fmt.Errorf("given height (%d) too high", height)
	}

	return bc.headers[height], nil
}
