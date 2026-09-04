package core

import (
	"bytes"
	"testing"

	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_Sign(t *testing.T) {
	privateKey := crypto.GeneratePrivateKey()
	tx := &Transaction{
		Data: []byte("test transaction"),
	}

	assert.Nil(t, tx.Sign(privateKey))
}	

func TestTransaction_Verify(t *testing.T) {
	privateKey := crypto.GeneratePrivateKey()
	tx := &Transaction{
		Data: []byte("test transaction"),
	}

	assert.Nil(t, tx.Sign(privateKey))
	assert.Nil(t, tx.Verify())

	otherPrivKey := crypto.GeneratePrivateKey()
	tx.From = otherPrivKey.PublicKey()

	assert.NotNil(t, tx.Verify())
}

func TestRransaction_Encode_Decode(t *testing.T) {
	tx := randomTxWithSignature(t)
	buf := &bytes.Buffer{}
	assert.Nil(t, tx.Encode(NewGobTxEncoder(buf)))

	txDecode := new(Transaction)
	assert.Nil(t, txDecode.Decode(NewGobTxDecoder(buf)))
	assert.Equal(t, tx, txDecode)
}

func randomTxWithSignature(t *testing.T) *Transaction {
	privKey := crypto.GeneratePrivateKey()

	tx := &Transaction{
		Data: []byte("random transaction"),
	}

	assert.Nil(t, tx.Sign(privKey))
	return tx
}