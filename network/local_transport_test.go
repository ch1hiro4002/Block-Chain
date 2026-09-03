package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	tra := NewLocalTransport("A")
	trb := NewLocalTransport("B")

	tra.Connect(trb)
	trb.Connect(tra)

	assert.Equal(t, tra.(*LocalTransport).peers["B"], trb)
	assert.Equal(t, trb.(*LocalTransport).peers["A"], tra)
}

func TestSendMessage(t *testing.T) {
	tra := NewLocalTransport("A")
	trb := NewLocalTransport("B")

	tra.Connect(trb)
	trb.Connect(tra)

	msg := []byte("Hello from A to B")

	go func() {
		rpc := <-trb.Consume()
		assert.Equal(t, rpc.Payload, msg)
		assert.Equal(t, rpc.From, tra.Addr())
	}()

	assert.Nil(t, tra.SendMessage("B", msg))
}
