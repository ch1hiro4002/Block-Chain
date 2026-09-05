package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"

	"github.com/ch1hiro4002/Block-Chain/core"
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
		Data: data,
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

type RPCHandler interface {
	HandleRPC(rpc RPC) error
}

type RPCProcessor interface {
	ProcessTransaction(NetAddr, *core.Transaction) error 
}

type DefaultRPCHandler struct {
	p RPCProcessor
}

func NewDefaultRPCHandler(p RPCProcessor) *DefaultRPCHandler {
	return &DefaultRPCHandler{
		p: p,
	}
}

func (h *DefaultRPCHandler) HandleRPC(rpc RPC) error {
	msg := Message{}
	if err := gob.NewDecoder(rpc.Payload).Decode(&msg); err != nil {
		return fmt.Errorf("failed to decode RPC payload: %v", err)
	}

	switch msg.Header {
	case MessageTypeTx:
		tx := new(core.Transaction)
		if err := tx.Decode(core.NewGobTxDecoder(bytes.NewReader(msg.Data))); err != nil {
			return fmt.Errorf("failed to decode TX data: %v", err)
		}

		return h.p.ProcessTransaction(rpc.From, tx)
	}

	return fmt.Errorf("invalid message header %d", msg.Header)
}
