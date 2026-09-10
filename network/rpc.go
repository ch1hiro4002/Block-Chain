package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"

	"github.com/ch1hiro4002/Block-Chain/core"
	"github.com/sirupsen/logrus"
)

type MessageType byte

const (
	MessageTypeTx        MessageType = 0x1
	MessageTypeBlock     MessageType = 0x2
	MessageTypeGetStatus MessageType = 0x3
	MessageTypeStatus    MessageType = 0x4
	MessageTypeGetBlocks MessageType = 0x5
	MessageTypeBlocks    MessageType = 0x6
)

type RPC struct {
	From    NetAddr
	Payload io.Reader
}

type Message struct {
	Header MessageType
	Data   []byte
}

func NewMessage(mt MessageType, data []byte) *Message {
	return &Message{
		Header: mt,
		Data:   data,
	}
}

func (msg *Message) Bytes() []byte {
	buf := &bytes.Buffer{}
	err := gob.NewEncoder(buf).Encode(msg)
	if err != nil {
		fmt.Printf("failed to encode message: %v", err)
		return nil
	}

	return buf.Bytes()
}

type RPCProcessor interface {
	ProcessMessage(*DecodeMessage) error
}

type RPCDecodeFunc func(RPC) (*DecodeMessage, error)

type DecodeMessage struct {
	From NetAddr
	Data any
}

func DefaultRPCDecoeFunc(rpc RPC) (*DecodeMessage, error) {
	msg := Message{}
	if err := gob.NewDecoder(rpc.Payload).Decode(&msg); err != nil {
		return nil, fmt.Errorf("failed to decode RPC payload: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"type": msg.Header,
		"from": rpc.From,
	}).Debug("new incoming message")

	switch msg.Header {
	case MessageTypeTx:
		tx := new(core.Transaction)
		if err := tx.Decode(core.NewGobTxDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, fmt.Errorf("failed to decode TX data: %v", err)
		}

		return &DecodeMessage{
			From: rpc.From,
			Data: tx,
		}, nil

	case MessageTypeBlock:
		block := new(core.Block)
		if err := block.Decode(core.NewGobBlockDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, fmt.Errorf("failed to decode Block data: %v", err)
		}

		return &DecodeMessage{
			From: rpc.From,
			Data: block,
		}, nil

	case MessageTypeGetStatus:
		return &DecodeMessage{
			From: rpc.From,
			Data: &GetStatusMessage{},
		}, nil

	case MessageTypeStatus:
		statusMessage := new(StatusMessage)
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(statusMessage); err != nil {
			return nil, fmt.Errorf("failed to decode Status data: %v", err)
		}

		return &DecodeMessage{
			From: rpc.From,
			Data: statusMessage,
		}, nil

	case MessageTypeGetBlocks:
		getBlocksMessage := new(GetBlocksMessage)
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(getBlocksMessage); err != nil {
			return nil, fmt.Errorf("failed to decode Status data: %v", err)
		}

		return &DecodeMessage{
			From: rpc.From,
			Data: getBlocksMessage,
		}, nil

	case MessageTypeBlocks:
		blocksMessage := new(BlocksMessage)
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(blocksMessage); err != nil {
			return nil, fmt.Errorf("failed to decode Status data: %v", err)
		}

		return &DecodeMessage{
			From: rpc.From,
			Data: blocksMessage,
		}, nil

	default:
		return nil, fmt.Errorf("invalid message header %v", msg.Header)
	}
}
