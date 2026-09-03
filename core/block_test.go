package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ch1hiro4002/Block-Chain/types"
	"github.com/ch1hiro4002/Block-Chain/crypto"
)
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

func randomBlock(height uint32, prevBlockHash types.Hash) *Block {
	header := &Header{
		Version:      1,
		PrevBlockHash: prevBlockHash,
		Timestamp:    time.Now().Unix(),
		Height:       height,
	}

	return NewBlock(header, []*Transaction{})
}

func randomBlockWithSignature(t *testing.T, height uint32, prevBlockHash types.Hash) *Block {
	privKey := crypto.GeneratePrivateKey()
	b := randomBlock(height, prevBlockHash)
	tx := randomTxWithSignature(t)
	b.AddTransaction(tx)

	assert.Nil(t, b.Sign(privKey))

	return b
}
