package core

import (
	"testing"

	"github.com/ch1hiro4002/Block-Chain/types"
	"github.com/stretchr/testify/assert"
)

func newBlockchainWithGenesis(t *testing.T) *Blockchain {
	bc, err := NewBlockchain(randomBlock(0, types.RandomHash()))

	assert.Nil(t, err)
	assert.Equal(t, bc.Height(), uint32(0))

	return bc
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

	lenBlocks := 100
	for i := 0; i < lenBlocks; i++ {
		block := randomBlockWithSignature(t, uint32(i + 1))
		bc.AddBlock(block)
	}

	assert.Equal(t, bc.Height(), uint32(lenBlocks))
	assert.Equal(t, len(bc.headers), lenBlocks+1)

	assert.NotNil(t, bc.AddBlock(randomBlock(56, types.RandomHash())))
}