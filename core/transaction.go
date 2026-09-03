package core

import (
	"fmt"

	"github.com/ch1hiro4002/Block-Chain/crypto"
)

type Transaction struct {
	Data []byte
	PublicKey crypto.PublicKey
	Signature *crypto.Signature
}

func (tx *Transaction) Sign(privateKey crypto.PrivateKey) error {
	sig, err := privateKey.Sign(tx.Data)
	if err != nil {
		return err
	}

	tx.Signature = sig
	tx.PublicKey = privateKey.PublicKey()

	return nil
}

func (tx *Transaction) Verify() error {
	if tx.Signature == nil {
		return fmt.Errorf("this transaction has no signature")
	}

	if !tx.Signature.Verify(tx.PublicKey, tx.Data) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}
