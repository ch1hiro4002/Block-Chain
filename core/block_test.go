package core

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/ch1hiro4002/Block-Chain/types"
	"github.com/stretchr/testify/assert"
)

func TestBlock_Hash(t *testing.T) {
	PrevBlockHash := types.Hash{}
	for i := 0; i < 5; i++ {
		block := randomBlock(t, uint32(i), PrevBlockHash)
		PrevBlockHash = block.PrevBlockHash
		fmt.Println(block.Hash(BlockHasher{}))
	}
}

func TestBlock_Sign(t *testing.T) {
	privateKey := crypto.GeneratePrivateKey()
	block := randomBlock(t, 66, types.Hash{})

	assert.Nil(t, block.Sign(privateKey))
}

func TestBlock_Verify(t *testing.T) {
	privateKey := crypto.GeneratePrivateKey()
	block := randomBlock(t, 66, types.Hash{})

	assert.Nil(t, block.Sign(privateKey))
	assert.Nil(t, block.Verify())

	otherPrivKey := crypto.GeneratePrivateKey()
	block.Validator = otherPrivKey.PublicKey()

	assert.NotNil(t, block.Verify())
}

func TestBlock_Encode_Decode(t *testing.T) {
	block := randomBlock(t, 455, types.Hash{})
	buf := &bytes.Buffer{}
	assert.Nil(t, block.Encode(NewGobBlockEncoder(buf)))

	bDecode := new(Block)
	assert.Nil(t, bDecode.Decode(NewGobBlockDecoder(buf)))

	assert.Equal(t, block, bDecode)
}

func randomBlock(t *testing.T, height uint32, prevBlockHash types.Hash) *Block {
	privKey := crypto.GeneratePrivateKey()

	header := &Header{
		Version:       1,
		TxHash:        types.Hash{},
		PrevBlockHash: prevBlockHash,
		Timestamp:     time.Now().Unix(),
		Height:        height,
	}
	block, err := NewBlock(header, []*Transaction{})
	assert.Nil(t, err)

	tx := randomTxWithSignature(t)
	block.AddTransaction(tx)

	dataHash, err := CalculateDataHash(block.Transactions)
	assert.Nil(t, err)
	block.TxHash = dataHash

	assert.Nil(t, block.Sign(privKey))

	return block
}
