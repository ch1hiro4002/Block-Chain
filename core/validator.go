package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/ch1hiro4002/Block-Chain/types"
)

var (
	ErrBlockAlreadyExists = errors.New("block already exists")
	ErrBlockTooHigh       = errors.New("block too high")
)

type Validator interface {
	ValidateBlock(b *Block) error
}

type BlockValidator struct {
	Address    types.Address
	Stake      uint64
	publicKey  crypto.PublicKey
	blockchain *BlockChain
}

func NewBlockValidator(addr types.Address, pubKey crypto.PublicKey, stake uint64, bc *BlockChain) *BlockValidator {
	return &BlockValidator{
		Address:    addr,
		Stake:      stake,
		publicKey:  pubKey,
		blockchain: bc,
	}
}

func (bv *BlockValidator) ValidateBlock(b *Block) error {
	if b.Height <= bv.blockchain.heightLocked() {
		return fmt.Errorf(
			"%w: height=%d hash=%s",
			ErrBlockAlreadyExists,
			b.Height,
			b.Hash(BlockHasher{}),
		)
	}

	if b.Height != bv.blockchain.heightLocked()+1 {
		return fmt.Errorf(
			"%w: height=%d current=%d hash=%s",
			ErrBlockTooHigh,
			b.Height,
			bv.blockchain.heightLocked(),
			b.Hash(BlockHasher{}),
		)
	}

	prevHeader, err := bv.blockchain.getHeaderLocked(b.Height - 1)
	if err != nil {
		return err
	}

	if err := b.Verify(prevHeader); err != nil {
		return fmt.Errorf("block verification failed: %v", err)
	}

	if vs := bv.blockchain.validatorSetForHeightLocked(b.Height); vs != nil && vs.Len() > 0 {
		expected := vs.Proposer(b.Height, b.PrevBlockHash)

		if b.Header.Proposer != expected {
			return fmt.Errorf("unexpected proposer: got %s, want %s", b.Header.Proposer, expected)
		}
	}

	return nil
}

type ValidatorInfo struct {
	Address   types.Address
	PublicKey crypto.PublicKey
	Stake     uint64
}

type ValidatorSet struct {
	mu         sync.RWMutex
	validators map[types.Address]BlockValidator
	order      []types.Address
	totalStake uint64
}

func NewValidatorSet() *ValidatorSet {
	return &ValidatorSet{
		validators: make(map[types.Address]BlockValidator),
	}
}

func (vs *ValidatorSet) Add(v BlockValidator) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	vs.addLocked(v)
	vs.normalizeLocked()
}

func (vs *ValidatorSet) addLocked(v BlockValidator) {
	if _, ok := vs.validators[v.Address]; !ok {
		vs.order = append(vs.order, v.Address)
	}

	vs.validators[v.Address] = v
}

func (vs *ValidatorSet) normalizeLocked() {
	sort.Slice(vs.order, func(i, j int) bool {
		return vs.order[i].String() < vs.order[j].String()
	})

	vs.totalStake = 0
	for _, addr := range vs.order {
		vs.totalStake += vs.validators[addr].Stake
	}
}

func (vs *ValidatorSet) Len() int {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	return len(vs.order)
}

func (vs *ValidatorSet) Clone() *ValidatorSet {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	clone := NewValidatorSet()
	clone.order = append([]types.Address(nil), vs.order...)
	clone.totalStake = vs.totalStake
	for addr, validator := range vs.validators {
		clone.validators[addr] = validator
	}

	return clone
}

func (vs *ValidatorSet) Proposer(height uint32, prevHash types.Hash) types.Address {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	if vs.totalStake == 0 {
		return types.Address{}
	}

	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, height)
	buf.Write(prevHash[:])

	seed := sha256.Sum256(buf.Bytes())
	target := binary.BigEndian.Uint64(seed[:8]) % vs.totalStake

	var cumulative uint64
	for _, addr := range vs.order {
		cumulative += vs.validators[addr].Stake
		if target < cumulative {
			return addr
		}
	}

	return types.Address{}
}

func (vs *ValidatorSet) Snapshot() []ValidatorInfo {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	snapshot := make([]ValidatorInfo, 0, len(vs.order))
	for _, addr := range vs.order {
		validator, ok := vs.validators[addr]
		if !ok {
			continue
		}

		snapshot = append(snapshot, ValidatorInfo{
			Address:   validator.Address,
			PublicKey: validator.publicKey,
			Stake:     validator.Stake,
		})
	}

	return snapshot
}

func (vs *ValidatorSet) Merge(infos []ValidatorInfo) error {
	if len(infos) == 0 {
		return nil
	}

	validated := make([]BlockValidator, 0, len(infos))
	seen := make(map[types.Address]struct{}, len(infos))

	for _, info := range infos {
		if err := validateValidatorInfo(info); err != nil {
			return err
		}

		if _, ok := seen[info.Address]; ok {
			return fmt.Errorf("duplicate validator address: %s", info.Address)
		}
		seen[info.Address] = struct{}{}

		validated = append(validated, BlockValidator{
			Address:   info.Address,
			Stake:     info.Stake,
			publicKey: info.PublicKey,
		})
	}

	vs.mu.Lock()
	defer vs.mu.Unlock()

	for _, validator := range validated {
		vs.addLocked(validator)
	}
	vs.normalizeLocked()

	return nil
}

func validateValidatorInfo(info ValidatorInfo) error {
	if info.Stake == 0 {
		return fmt.Errorf("validator %s has zero stake", info.Address)
	}

	publicKeyBytes, err := info.PublicKey.Marshal()
	if err != nil {
		return fmt.Errorf("validator %s has invalid public key: %w", info.Address, err)
	}
	if len(publicKeyBytes) == 0 {
		return fmt.Errorf("validator %s has empty public key", info.Address)
	}

	if info.PublicKey.Address() != info.Address {
		return fmt.Errorf("validator public key does not match address: address=%s", info.Address)
	}

	return nil
}
