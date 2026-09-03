package types

import (
	"encoding/hex"
	"fmt"
)

type Address [20]byte

func (a Address) String() string {
	return hex.EncodeToString(a[:])
}

func AddressFromBytes(b []byte) Address {
	if len(b) != 20 {
		err := fmt.Sprintf("given bytes with length %d, expected 20", len(b))
		panic(err)
	}

	return Address(b)
}
