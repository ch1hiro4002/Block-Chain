package core

import (
	"testing"

	"github.com/ch1hiro4002/Block-Chain/types"
	"github.com/stretchr/testify/assert"
)

func newBlockchainWithGenesis(t *testing.T) *Blockchain {
	bc, err := NewBlockchain(randomBlock(0, types.Hash{}))

	assert.Nil(t, err)
	assert.Equal(t, bc.Height(), uint32(0))

	return bc
}

func getPrevBlockHash(t *testing.T, bc *Blockchain, height uint32) types.Hash {
	header, err := bc.GetHeader(height - 1)
	assert.Nil(t, err)

	return BlockHasher{}.Hash(header)
}

func TestNewBlockchain(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	assert.NotNil(t, bc)
}

func TestBlockchain_HasBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	assert.True(t, bc.HasBlock(0))
}

func TestBlockchain_AddBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)

	lenBlocks := 5
	for i := 0; i < lenBlocks; i++ {
		block := randomBlockWithSignature(t, uint32(i + 1), getPrevBlockHash(t, bc, uint32(i + 1)))
		bc.AddBlock(block)
	}

	assert.Equal(t, bc.Height(), uint32(lenBlocks))
	assert.Equal(t, len(bc.headers), lenBlocks+1)

	// Adding a duplicate block should return an error
	assert.NotNil(t, bc.AddBlock(randomBlockWithSignature(t, uint32(lenBlocks), getPrevBlockHash(t, bc, uint32(lenBlocks)))))

	// Adding a block with a height that is too high should return an error
	assert.NotNil(t, bc.AddBlock(randomBlockWithSignature(t, 102, types.RandomHash())))
}