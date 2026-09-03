package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ch1hiro4002/Block-Chain/types"
	"github.com/ch1hiro4002/Block-Chain/crypto"
)

func randomBlock(height uint32, PrevBlockHash types.Hash) *Block {
	header := &Header{
		Version:      1,
		PrevBlockHash: PrevBlockHash,
		Timestamp:    time.Now().Unix(),
		Height:       height,
	}

	tx := Transaction{
		Data: []byte("random transaction"),
	}

	return NewBlock(header, []Transaction{tx})
}

func randomBlockWithSignature(t *testing.T, height uint32) *Block {
	privKey := crypto.GeneratePrivateKey()
	b := randomBlock(height, types.RandomHash())

	assert.Nil(t, b.Sign(privKey))
	
	return b
}

func TestBlock_Hash(t *testing.T) {
	PrevBlockHash := types.RandomHash()
	for i := 0; i < 5; i++ {
		block := randomBlock(uint32(i), PrevBlockHash)
		PrevBlockHash = block.PrevBlockHash
		fmt.Println(block.Hash(BlockHasher{}))
	}
}

func TestBlock_Sign(t *testing.T) {
	privateKey := crypto.GeneratePrivateKey()
	block := randomBlock(66, types.RandomHash())

	assert.Nil(t, block.Sign(privateKey))
}

func TestBlock_Verify(t *testing.T) {
	privateKey := crypto.GeneratePrivateKey()
	block := randomBlock(66, types.RandomHash())

	assert.Nil(t, block.Sign(privateKey))
	assert.Nil(t, block.Verify())

	otherPrivKey := crypto.GeneratePrivateKey()
	block.Validator = otherPrivKey.PublicKey()

	assert.NotNil(t, block.Verify())
}