package core

import "fmt"

type Storage interface {
	Put(*Block) error
	Get(uint32) (*Block, error)
	GetRange(uint32, uint32) ([]*Block, error)
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

func (s *MemoryStore) GetRange(from, to uint32) ([]*Block, error) {
	if from > to {
		return nil, fmt.Errorf("invalid block range: from (%d) > to (%d)", from, to)
	}

	blocks := make([]*Block, 0, to-from+1)

	for height := from; height <= to; height++ {
		block, ok := s.blocks[height]
		if !ok {
			return nil, fmt.Errorf("block not found at height %d", height)
		}

		blocks = append(blocks, block)
	}

	return blocks, nil
}
