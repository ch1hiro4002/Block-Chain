package network

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransport_Connect(t *testing.T) {
	tra := NewLocalTransport("A")
	trb := NewLocalTransport("B")

	tra.Connect(trb)
	trb.Connect(tra)

	assert.Equal(t, tra.(*LocalTransport).peers["B"], trb)
	assert.Equal(t, trb.(*LocalTransport).peers["A"], tra)
}

func TestTransport_SendMessage(t *testing.T) {
	tra := NewLocalTransport("A")
	trb := NewLocalTransport("B")

	tra.Connect(trb)
	trb.Connect(tra)

	msg := []byte("Hello from A to B")
	assert.Nil(t, tra.SendMessage(trb.Addr(), msg))

	rpc := <-trb.Consume()
	b, err := io.ReadAll(rpc.Payload)
	assert.Nil(t, err)
	assert.Equal(t, msg, b)
	assert.Equal(t, rpc.From, tra.Addr())
}

func TestTransport_Broadcast(t *testing.T) {
	tra := NewLocalTransport("A")
	trb := NewLocalTransport("B")
	trc := NewLocalTransport("C")

	tra.Connect(trb)
	tra.Connect(trc)

	msg := []byte("Transaction_A")
	assert.Nil(t, tra.Broadcast(msg))

	rpcb := <-trb.Consume()
	b, err := io.ReadAll(rpcb.Payload)
	assert.Nil(t, err)
	assert.Equal(t, msg, b)

	rpcc := <-trc.Consume()
	c, err := io.ReadAll(rpcc.Payload)
	assert.Nil(t, err)
	assert.Equal(t, msg, c)
}
