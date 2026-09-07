package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/ch1hiro4002/Block-Chain/core"
	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/ch1hiro4002/Block-Chain/network"
)

func main() {
	trLocal := network.NewLocalTransport("LOCAL")
	trRemoteA := network.NewLocalTransport("REMOTE_A")
	trRemoteB := network.NewLocalTransport("REMOTE_A")
	trRemoteC := network.NewLocalTransport("REMOTE_A")

	trLocal.Connect(trRemoteA)
	trRemoteA.Connect(trRemoteB)
	trRemoteB.Connect(trRemoteC)

	trRemoteA.Connect(trLocal)

	initRemoteServers([]network.Transport{trRemoteA, trRemoteB, trRemoteC})

	go func() {
		for {
			err := trRemoteA.SendMessage(trLocal.Addr(), createTxMessage())
			if err != nil {
				log.Fatal(err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	privKey := crypto.GeneratePrivateKey()

	localServer := makeServer("LOCAL", &privKey, trLocal)

	localServer.Strat()
}

func initRemoteServers(trs []network.Transport) {
	for i := 0; i < len(trs); i++ {
		id := fmt.Sprintf("REMOTE_%d", i)
		s := makeServer(id, nil, trs[i])
		go s.Strat()
	}
}

func makeServer(id string, privKey *crypto.PrivateKey, tr network.Transport) *network.Server {
	opts := network.ServerOpts{
		ID:         id,
		PrivateKey: privKey,
		Transports: []network.Transport{tr},
	}

	s, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}

	return s
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
