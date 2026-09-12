package core

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/ch1hiro4002/Block-Chain/types"
)

type Transaction struct {
	Nonce     uint64
	Timestamp int64
	Deadline  int64
	Data      []byte
	From      crypto.PublicKey
	Signature *crypto.Signature
	hash      types.Hash
}

func NewTransaction(nonce uint64, timestamp int64, deadline int64, data []byte) *Transaction {
	return &Transaction{
		Nonce:     nonce,
		Timestamp: timestamp,
		Deadline:  deadline,
		Data:      data,
	}
}

func (tx *Transaction) Hash(hasher Hasher[*Transaction]) types.Hash {
	if tx.hash.IsZero() {
		tx.hash = hasher.Hash(tx)
	}

	return tx.hash
}

func (tx *Transaction) Sign(privateKey crypto.PrivateKey) error {
	tx.From = privateKey.PublicKey()

	sig, err := privateKey.Sign(tx.Hash(TxHasher{}).Bytes())
	if err != nil {
		return err
	}

	tx.Signature = sig

	return nil
}

func (tx *Transaction) Verify() error {
	if tx.Signature == nil {
		return fmt.Errorf("this transaction has no signature")
	}

	if !tx.Signature.Verify(tx.From, tx.Hash(TxHasher{}).Bytes()) {
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

func (tx *Transaction) signableBytes() []byte {
	var buf bytes.Buffer

	buf.WriteString("blockchain-tx:v1")
	buf.WriteByte(0x00)

	writeUint64(&buf, tx.Nonce)
	writeUint64(&buf, uint64(tx.Timestamp))
	writeUint64(&buf, uint64(tx.Deadline))

	fromBytes, _ := tx.From.Marshal()
	writeBytesWithLength(&buf, fromBytes)
	writeBytesWithLength(&buf, tx.Data)

	return buf.Bytes()
}

func writeUint64(buf *bytes.Buffer, v uint64) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], v)
	buf.Write(b[:])
}
func writeBytesWithLength(buf *bytes.Buffer, data []byte) {
	writeUint64(buf, uint64(len(data)))
	buf.Write(data)
}
