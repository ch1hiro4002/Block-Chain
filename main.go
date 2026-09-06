package main

import (
	"bytes"
	"crypto/rand"
	"time"

	"github.com/ch1hiro4002/Block-Chain/core"
	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/ch1hiro4002/Block-Chain/network"
)

func main() {
	trLocal := network.NewLocalTransport("LOCAL")
	trRemote := network.NewLocalTransport("REMOTE")

	trLocal.Connect(trRemote)
	trRemote.Connect(trLocal)

	go func() {
		for {
			trRemote.SendMessage(trLocal.Addr(), createTxMessage())
			time.Sleep(3 * time.Second)
		}
	}()

	opts := network.ServerOpts{
		Transports: []network.Transport{trLocal},
	}

	s := network.NewServer(opts)
	s.Strat()
}

func createTxMessage() []byte {
	privKey := crypto.GeneratePrivateKey()

	data := make([]byte, 32)
	rand.Read(data)

	tx := core.NewTransaction(data)
	tx.Sign(privKey)

	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return nil
	}

	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

	return msg.Bytes()
}
