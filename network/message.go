package network

import (
	"bytes"
	"encoding/gob"

	"github.com/ch1hiro4002/Block-Chain/core"
)

type GetStatusMessage struct{}

type StatusMessage struct {
	ID            string
	CurrentHeight uint32
}

type GetBlocksMessage struct {
	From uint32
	// if to is 0 the maximum blocks will be returned
	To uint32
}

type BlocksMessage struct {
	Blocks []*core.Block
}

func (bm *BlocksMessage) GobEncode() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)

	if err := enc.Encode(len(bm.Blocks)); err != nil {
		return nil, err
	}

	for _, block := range bm.Blocks {
		var blockBuf bytes.Buffer

		if err := block.Encode(core.NewGobBlockEncoder(&blockBuf)); err != nil {
			return nil, err
		}

		if err := enc.Encode(blockBuf.Bytes()); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func (m *BlocksMessage) GobDecode(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))

	var count int
	if err := dec.Decode(&count); err != nil {
		return err
	}

	m.Blocks = make([]*core.Block, 0, count)

	for i := 0; i < count; i++ {
		var blockBytes []byte
		if err := dec.Decode(&blockBytes); err != nil {
			return err
		}

		block := new(core.Block)
		if err := block.Decode(core.NewGobBlockDecoder(bytes.NewReader(blockBytes))); err != nil {
			return err
		}

		m.Blocks = append(m.Blocks, block)
	}

	return nil
}
