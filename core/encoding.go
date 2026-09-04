package core

import (
	"encoding/gob"
	"io"

	"github.com/ch1hiro4002/Block-Chain/crypto"
)

type Encoder[T any] interface {
	Encode(T) error
}

type Decoder[T any] interface {
	Decode(T) error
}

type GobTxEncoder struct {
	w io.Writer
}

func NewGobTxEncoder(w io.Writer) *GobTxEncoder {
	return &GobTxEncoder{
		w: w,
	}
}

func (e *GobTxEncoder) Encode(tx *Transaction) error {
	from, err := tx.From.Marshal()
	if err != nil {
		return err
	}

	var sig []byte
	if tx.Signature != nil {
		sig, err = tx.Signature.Marshal()
		if err != nil {
			return err
		}
	}

	w := struct {
		Data      []byte
		From      []byte
		Signature []byte
	}{
		Data:      tx.Data,
		From:      from,
		Signature: sig,
	}

	return gob.NewEncoder(e.w).Encode(w)
}

type GobTxDecoder struct {
	r io.Reader
}

func NewGobTxDecoder(r io.Reader) *GobTxDecoder {
	return &GobTxDecoder{
		r: r,
	}
}

func (d *GobTxDecoder) Decode(tx *Transaction) error {
	w := struct {
		Data      []byte
		From      []byte
		Signature []byte
	}{}

	if err := gob.NewDecoder(d.r).Decode(&w); err != nil {
		return err
	}

	from, err := crypto.UnmarshalPublicKey(w.From)
	if err != nil {
		return err
	}

	tx.Data = w.Data
	tx.From = from

	if len(w.Signature) > 0 {
		sig, err := crypto.UnmarshalSignature(w.Signature)
		if err != nil {
			return err
		}
		tx.Signature = &sig
	}

	return nil
}
