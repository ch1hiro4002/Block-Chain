package main

import (
	"bytes"
	"crypto/rand"
	"time"

	"github.com/ch1hiro4002/Block-Chain/core"
	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/ch1hiro4002/Block-Chain/network"
	"github.com/sirupsen/logrus"
)

func main() {
	trLocal := network.NewLocalTransport("LOCAL")
	trRemote := network.NewLocalTransport("REMOTE")

	trLocal.Connect(trRemote)
	trRemote.Connect(trLocal)

	go func() {
		for {
			payload, err := createTxByte()
			if err != nil {
				logrus.Error(err)
			}
			trRemote.SendMessage(trLocal.Addr(), network.MessageTypeTx, payload)
			time.Sleep(3 * time.Second)
		}
	}()

	opts := network.ServerOpts{
		Transports: []network.Transport{trLocal},
	}

	s := network.NewServer(opts)
	s.Strat()
}

func createTxByte() ([]byte, error) {
	privKey := crypto.GeneratePrivateKey()
	data := make([]byte, 32)
	rand.Read(data)
	tx := core.NewTransaction(data)
	tx.Sign(privKey)
	tx.Hash(core.TxHasher{})
	tx.SetTime(time.Now())

	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

