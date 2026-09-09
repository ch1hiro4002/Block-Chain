package core

import "fmt"

type Storage interface {
	Put(*Block) error
	Get(uint32) (*Block, error)
}

type MemoryStore struct {
	blocks map[uint32]*Block
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		blocks: make(map[uint32]*Block),
	}
}

func (s *MemoryStore) Put(b *Block) error {
	s.blocks[b.Height] = b
	return nil
}

func (s *MemoryStore) Get(height uint32) (*Block, error) {
	block, ok := s.blocks[height]
	if !ok {
		return nil, fmt.Errorf("block not found at height %d", height)
	}

	return block, nil
}