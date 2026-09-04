package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/gob"
	"fmt"
	"math/big"

	"github.com/ch1hiro4002/Block-Chain/types"
)

type PrivateKey struct {
	key *ecdsa.PrivateKey
}

type PublicKey struct {
	key *ecdsa.PublicKey
}

type Signature struct {
	r, s *big.Int
}

// Marshal serializes the public key into PKIX ASN.1 DER bytes.
func (k PublicKey) Marshal() ([]byte, error) {
	if k.key == nil {
		return nil, nil
	}
	return x509.MarshalPKIXPublicKey(k.key)
}

// UnmarshalPublicKey parses PKIX ASN.1 DER bytes into a PublicKey.
func UnmarshalPublicKey(data []byte) (PublicKey, error) {
	if len(data) == 0 {
		return PublicKey{}, nil
	}

	pub, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return PublicKey{}, err
	}

	ec, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return PublicKey{}, fmt.Errorf("unsupported public key type: %T", pub)
	}

	return PublicKey{key: ec}, nil
}

type signatureWire struct {
	R []byte
	S []byte
}

// Marshal serializes the signature's r and s values into bytes.
func (sig Signature) Marshal() ([]byte, error) {
	w := signatureWire{}
	if sig.r != nil {
		w.R = sig.r.Bytes()
	}
	if sig.s != nil {
		w.S = sig.s.Bytes()
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(w); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalSignature parses bytes into a Signature.
func UnmarshalSignature(data []byte) (Signature, error) {
	var w signatureWire
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&w); err != nil {
		return Signature{}, err
	}

	return Signature{
		r: new(big.Int).SetBytes(w.R),
		s: new(big.Int).SetBytes(w.S),
	}, nil
}

func GeneratePrivateKey() PrivateKey {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	return PrivateKey{
		key: key,
	}
}

func (k PrivateKey) PublicKey() PublicKey {
	return PublicKey{
		key: &k.key.PublicKey,
	}
}

func (k PublicKey) ToSlice() []byte {
	bytes, err := k.key.Bytes()
	if err != nil {
		panic(err)
	}

	return bytes
}

func (k PublicKey) Address() types.Address {
	h := sha256.Sum256(k.ToSlice())

	return types.AddressFromBytes(h[len(h)-20:])
}

func (k PrivateKey) Sign(data []byte) (*Signature, error) {
	r, s, err := ecdsa.Sign(rand.Reader, k.key, data)
	if err != nil {
		return nil, err
	}

	return &Signature{r, s}, nil
}

func (sig Signature) Verify(pubKey PublicKey, data []byte) bool {
	return ecdsa.Verify(pubKey.key, data, sig.r, sig.s)
}
