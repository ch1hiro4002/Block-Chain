package types

import (
	"crypto/rand"
	"encoding/hex"
)


type Hash [32]byte

func (h Hash) IsZero() bool {
	for _, v := range h {
		if v != 0 {
			return false
		}
	}
	return true
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}


func RandomBytes(size int) []byte {
	token := make([]byte, size)
	rand.Read(token)
	return token
}

func RandomHash() Hash {
	return Hash(RandomBytes(32))
}
