package core

import (
	"bytes"
	"fmt"

	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/ch1hiro4002/Block-Chain/types"
)

type Transaction struct {
	Data      []byte
	From      crypto.PublicKey
	Signature *crypto.Signature
	hash      types.Hash
}

func NewTransaction(data []byte) *Transaction {
	return &Transaction{
		Data: data,
	}
}

func (tx *Transaction) Hash(hasher Hasher[*Transaction]) types.Hash {
	if tx.hash.IsZero() {
		tx.hash = hasher.Hash(tx)
	}

	return tx.hash
}

func (tx *Transaction) Bytes() []byte {
	buf := &bytes.Buffer{}
	tx.Encode(NewGobTxEncoder(buf))

	return buf.Bytes()
}

func (tx *Transaction) Sign(privateKey crypto.PrivateKey) error {
	sig, err := privateKey.Sign(tx.Data)
	if err != nil {
		return err
	}

	tx.Signature = sig
	tx.From = privateKey.PublicKey()

	return nil
}

func (tx *Transaction) Verify() error {
	if tx.Signature == nil {
		return fmt.Errorf("this transaction has no signature")
	}

	if !tx.Signature.Verify(tx.From, tx.Data) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func (tx *Transaction) Encode(enc Encoder[*Transaction]) error {
	return enc.Encode(tx)
}

func (tx *Transaction) Decode(dec Decoder[*Transaction]) error {
	return dec.Decode(tx)
}
