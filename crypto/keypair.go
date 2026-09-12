package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ch1hiro4002/Block-Chain/types"
)

type PrivateKey struct {
	key *ecdsa.PrivateKey
}

func (k PrivateKey) PublicKey() PublicKey {
	return PublicKey{
		key: &k.key.PublicKey,
	}
}

func (k PrivateKey) Sign(data []byte) (*Signature, error) {
	r, s, err := ecdsa.Sign(rand.Reader, k.key, data)
	if err != nil {
		return nil, err
	}

	return &Signature{r: r, s: s}, nil
}

type PublicKey struct {
	key *ecdsa.PublicKey
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

func (k PublicKey) GobEncode() ([]byte, error) {
	return k.Marshal()
}

func (k *PublicKey) GobDecode(data []byte) error {
	pub, err := UnmarshalPublicKey(data)
	if err != nil {
		return err
	}

	*k = pub
	return nil
}

type Signature struct {
	r, s *big.Int
}

func (sig Signature) R() *big.Int {
	if sig.r == nil {
		return nil
	}
	return new(big.Int).Set(sig.r)
}

func (sig Signature) S() *big.Int {
	if sig.s == nil {
		return nil
	}
	return new(big.Int).Set(sig.s)
}

func (sig Signature) GobEncode() ([]byte, error) {
	var buf bytes.Buffer

	writeSignatureInt(&buf, sig.r)
	writeSignatureInt(&buf, sig.s)

	return buf.Bytes(), nil
}

func (sig *Signature) GobDecode(data []byte) error {
	r, rest, err := readSignatureInt(data)
	if err != nil {
		return err
	}

	s, rest, err := readSignatureInt(rest)
	if err != nil {
		return err
	}

	if len(rest) != 0 {
		return fmt.Errorf("invalid signature encoding: trailing bytes")
	}

	sig.r = r
	sig.s = s

	return nil
}

func writeSignatureInt(buf *bytes.Buffer, v *big.Int) {
	if v == nil {
		buf.WriteByte(0)
		return
	}

	b := v.Bytes()
	buf.WriteByte(1)

	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(b)))
	buf.Write(length[:])
	buf.Write(b)
}

func readSignatureInt(data []byte) (*big.Int, []byte, error) {
	if len(data) == 0 {
		return nil, nil, fmt.Errorf("invalid signature encoding: missing value flag")
	}

	if data[0] == 0 {
		return nil, data[1:], nil
	}

	if len(data) < 9 {
		return nil, nil, fmt.Errorf("invalid signature encoding: truncated length")
	}

	length := binary.BigEndian.Uint64(data[1:9])
	if length > uint64(len(data)-9) {
		return nil, nil, fmt.Errorf("invalid signature encoding: length out of range")
	}

	start := 9
	end := start + int(length)
	value := new(big.Int).SetBytes(data[start:end])

	return value, data[end:], nil
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

func (k PublicKey) ToSlice() []byte {
	if k.key == nil {
		return nil
	}
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

func (sig Signature) Verify(pubKey PublicKey, data []byte) bool {
	return ecdsa.Verify(pubKey.key, data, sig.r, sig.s)
}
