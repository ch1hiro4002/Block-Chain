package core

import (
	"testing"

	"github.com/ch1hiro4002/Block-Chain/types"
	"github.com/stretchr/testify/assert"
)

func newBlockChainWithGenesis(t *testing.T) *BlockChain {
	bc, err := NewBlockChain()

	assert.Nil(t, err)
	assert.Equal(t, bc.Height(), uint32(0))

	return bc
}

func getPrevBlockHash(t *testing.T, bc *BlockChain, height uint32) types.Hash {
	header, err := bc.GetHeader(height - 1)
	assert.Nil(t, err)

	return BlockHasher{}.Hash(header)
}

func TestNewBlockChain(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	assert.NotNil(t, bc)
}

func TestBlockChain_HasBlock(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	assert.True(t, bc.HasBlock(0))
}

func TestBlockChain_AddBlock(t *testing.T) {
	bc := newBlockChainWithGenesis(t)

	lenBlocks := 5
	for i := 0; i < lenBlocks; i++ {
		block := randomBlock(t, uint32(i + 1), getPrevBlockHash(t, bc, uint32(i + 1)))
		bc.AddBlock(block)
	}

	assert.Equal(t, bc.Height(), uint32(lenBlocks))
	assert.Equal(t, len(bc.headers), lenBlocks+1)

	// Adding a duplicate block should return an error
	assert.NotNil(t, bc.AddBlock(randomBlock(t, uint32(lenBlocks), getPrevBlockHash(t, bc, uint32(lenBlocks)))))

	// Adding a block with a height that is too high should return an error
	assert.NotNil(t, bc.AddBlock(randomBlock(t, 102, types.RandomHash())))
}