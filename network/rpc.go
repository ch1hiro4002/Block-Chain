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
	MessageTypeTx MessageType = 0x1
	MessageTypeBock
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
	default:
		return nil, fmt.Errorf("invalid message header %v", msg.Header)
	}
}
