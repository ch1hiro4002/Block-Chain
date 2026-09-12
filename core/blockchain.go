package core

import (
	"fmt"
	"sync"
	"time"

	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/ch1hiro4002/Block-Chain/types"
	"github.com/go-kit/log"
)

const (
	MaxTransactionLifetime = 5 * time.Minute
)

type BlockChain struct {
	logger              log.Logger
	store               Storage
	headers             []*Header
	hashes              map[types.Hash]uint32
	validator           Validator
	lock                sync.RWMutex
	state               *State
	accountNonces       map[types.Address]uint64
	validatorSet        *ValidatorSet
	validatorSetHistory map[uint32]*ValidatorSet
}

func NewBlockChain(logger log.Logger, addr types.Address, pubKey crypto.PublicKey, stake uint64) (*BlockChain, error) {
	bc := &BlockChain{
		logger:              logger,
		store:               NewMemoryStore(),
		headers:             []*Header{},
		hashes:              make(map[types.Hash]uint32),
		state:               NewState(),
		accountNonces:       make(map[types.Address]uint64),
		validatorSetHistory: make(map[uint32]*ValidatorSet),
	}
	bv := NewBlockValidator(addr, pubKey, stake, bc)
	bc.validator = bv
	bc.validatorSet = NewValidatorSet()
	bc.validatorSet.Add(*bv)

	genesis, err := newGenesisBlock()
	if err != nil {
		return nil, err
	}
	if err := bc.addBlockWithoutValidation(genesis); err != nil {
		return nil, err
	}

	return bc, nil
}

func (bc *BlockChain) ValidateTransaction(tx *Transaction, now time.Time) error {
	if err := tx.Verify(); err != nil {
		return err
	}

	if err := validateTxDeadline(tx, now); err != nil {
		return err
	}

	bc.lock.RLock()
	defer bc.lock.RUnlock()

	currentNonce := bc.accountNonces[tx.From.Address()]
	if tx.Nonce != currentNonce {
		return fmt.Errorf("invalid nonce: got %d, want %d", tx.Nonce, currentNonce)
	}

	return nil
}

func (bc *BlockChain) AddBlock(block *Block) error {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	// Validate the block
	if err := bc.validator.ValidateBlock(block); err != nil {
		return err
	}

	for _, tx := range block.Transactions {
		if err := bc.executeTransactionLocked(tx, time.Now()); err != nil {
			return err
		}
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

func (bc *BlockChain) GetValidatorSet() *ValidatorSet {
	return bc.validatorSet
}

func (bc *BlockChain) GetValidatorSetForHeight(height uint32) *ValidatorSet {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	return bc.validatorSetForHeightLocked(height)
}

func (bc *BlockChain) validatorSetForHeightLocked(height uint32) *ValidatorSet {
	if validatorSet, ok := bc.validatorSetHistory[height]; ok {
		return validatorSet
	}

	return bc.validatorSet
}

func (bc *BlockChain) GetValidatorSetHistory() map[uint32][]ValidatorInfo {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	history := make(map[uint32][]ValidatorInfo, len(bc.validatorSetHistory))
	for height, validatorSet := range bc.validatorSetHistory {
		history[height] = validatorSet.Snapshot()
	}

	return history
}

func (bc *BlockChain) SetValidatorSetHistory(history map[uint32][]ValidatorInfo) error {
	pending := make(map[uint32]*ValidatorSet, len(history))
	for height, infos := range history {
		validatorSet := NewValidatorSet()
		if err := validatorSet.Merge(infos); err != nil {
			return err
		}
		pending[height] = validatorSet
	}

	bc.lock.Lock()
	defer bc.lock.Unlock()

	for height, validatorSet := range pending {
		if _, ok := bc.validatorSetHistory[height]; ok {
			continue
		}
		bc.validatorSetHistory[height] = validatorSet
	}

	return nil
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
		"proposer", b.Proposer,
	)

	if err := bc.store.Put(b); err != nil {
		return err
	}

	bc.recordValidatorSetLocked(b.Height)

	return nil
}

func (bc *BlockChain) recordValidatorSetLocked(height uint32) {
	if _, ok := bc.validatorSetHistory[height]; ok {
		return
	}

	bc.validatorSetHistory[height] = bc.validatorSet.Clone()
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

func (bc *BlockChain) executeTransactionLocked(tx *Transaction, now time.Time) error {
	if err := bc.validateTransactionLocked(tx, now); err != nil {
		return err
	}

	vm := NewVM(tx.Data, bc.state)
	if err := vm.Run(); err != nil {
		return err
	}

	bc.accountNonces[tx.From.Address()]++

	return nil
}

func (bc *BlockChain) validateTransactionLocked(tx *Transaction, now time.Time) error {
	if err := tx.Verify(); err != nil {
		return err
	}

	if err := validateTxDeadline(tx, now); err != nil {
		return err
	}

	expected := bc.accountNonces[tx.From.Address()]
	if tx.Nonce != expected {
		return fmt.Errorf(
			"invalid nonce: got %d, want %d",
			tx.Nonce,
			expected,
		)
	}

	return nil
}

func validateTxDeadline(tx *Transaction, now time.Time) error {
	nowUnix := now.Unix()

	if tx.Deadline == 0 {
		return fmt.Errorf("transaction has no deadline")
	}

	if tx.Deadline <= nowUnix {
		return fmt.Errorf("transaction has expired")
	}

	if tx.Deadline-nowUnix > int64(MaxTransactionLifetime/time.Second) {
		return fmt.Errorf("transaction deadline is too far in the future")
	}

	return nil
}
