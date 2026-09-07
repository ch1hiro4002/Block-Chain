package core

import (
	"errors"
	"fmt"
)

var (
	ErrBlockAlreadyExists = errors.New("block already exists")
	ErrBlockTooHigh       = errors.New("block too high")
)

type Validator interface {
	ValidateBlock(*Block) error
}

type BlockValidator struct {
	blockchain *BlockChain
}

func NewBlockValidator(bc *BlockChain) *BlockValidator {
	return &BlockValidator{
		blockchain: bc,
	}
}

func (bv *BlockValidator) ValidateBlock(b *Block) error {
	if b.Height <= bv.blockchain.heightLocked() {
		return fmt.Errorf(
			"%w: height=%d hash=%s",
			ErrBlockAlreadyExists,
			b.Height,
			b.Hash(BlockHasher{}),
		)
	}

	if b.Height != bv.blockchain.heightLocked()+1 {
		return fmt.Errorf(
			"%w: height=%d current=%d hash=%s",
			ErrBlockTooHigh,
			b.Height,
			bv.blockchain.heightLocked(),
			b.Hash(BlockHasher{}),
		)
	}

	prevHeader, err := bv.blockchain.getHeaderLocked(b.Height - 1)
	if err != nil {
		return err
	}

	hash := BlockHasher{}.Hash(prevHeader)
	if hash != b.PrevBlockHash {
		return fmt.Errorf("block(%d) has invalid previous block hash", b.Height)
	}

	if err := b.Verify(); err != nil {
		return fmt.Errorf("block verification failed: %v", err)
	}
	return nil
}
