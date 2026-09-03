package core

import (
	"fmt"
)

type Validator interface {
	ValidateBlock(*Block) error
}

type BlockValidator struct {
	bc *Blockchain
}

func NewBlockValidator(bc *Blockchain) *BlockValidator {
	return &BlockValidator{
		bc: bc,
	}
}

func (bv *BlockValidator) ValidateBlock(b *Block) error {
	if b.Height <= bv.bc.heightLocked() {
		return fmt.Errorf("block height %d is invalid", b.Height)
	}

	if b.Height != bv.bc.heightLocked()+1 {
		return fmt.Errorf("block(%d) too high", b.Height)
	}

	prevHeader, err := bv.bc.getHeaderLocked(b.Height - 1)
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
