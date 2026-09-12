package core

import (
	"testing"

	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/stretchr/testify/assert"
)

func TestValidatorSetMergeAndSnapshot(t *testing.T) {
	vs := NewValidatorSet()

	privKeyA := crypto.GeneratePrivateKey()
	pubKeyA := privKeyA.PublicKey()

	privKeyB := crypto.GeneratePrivateKey()
	pubKeyB := privKeyB.PublicKey()

	err := vs.Merge([]ValidatorInfo{
		{
			Address:   pubKeyA.Address(),
			PublicKey: pubKeyA,
			Stake:     100,
		},
		{
			Address:   pubKeyB.Address(),
			PublicKey: pubKeyB,
			Stake:     300,
		},
	})
	assert.NoError(t, err)

	snapshot := vs.Snapshot()
	assert.Len(t, snapshot, 2)

	stakes := make(map[string]uint64, len(snapshot))
	for _, validator := range snapshot {
		stakes[validator.Address.String()] = validator.Stake
	}

	assert.Equal(t, uint64(100), stakes[pubKeyA.Address().String()])
	assert.Equal(t, uint64(300), stakes[pubKeyB.Address().String()])
}

func TestValidatorSetMergeRejectsInvalidValidator(t *testing.T) {
	vs := NewValidatorSet()

	privKeyA := crypto.GeneratePrivateKey()
	pubKeyA := privKeyA.PublicKey()

	privKeyB := crypto.GeneratePrivateKey()
	pubKeyB := privKeyB.PublicKey()

	err := vs.Merge([]ValidatorInfo{
		{
			Address:   pubKeyA.Address(),
			PublicKey: pubKeyB,
			Stake:     100,
		},
	})
	assert.Error(t, err)
	assert.Empty(t, vs.Snapshot())
}
