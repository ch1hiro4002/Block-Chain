package main

import (
	"bytes"
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
			time.Sleep(2 * time.Second)
		}
	}()

	// go func() {
	// 	time.Sleep(6 * time.Second)

	// 	trLatest := network.NewLocalTransport("LATEST_REMOTE")
	// 	trRemoteC.Connect(trLatest)

	// 	latestServer := makeServer("LATEST_REMOTE", nil, trLatest)
	// 	latestServer.Strat()
	// } ()

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

func contract() []byte {
	fooData := []byte{0x46, 0x0b, 0x4f, 0x0b, 0x4f, 0x0b, 0x03, 0x0a, 0x10, 0x12}
	data := []byte{0x03, 0x0a, 0x02, 0x0a, 0x0c, 0x46, 0x0b, 0x4f, 0x0b, 0x4f, 0x0b, 0x03, 0x0a, 0x10, 0x11}
	data = append(data, fooData...)

	return data
}

func createTxMessage() []byte {
	privKey := crypto.GeneratePrivateKey()

	tx := core.NewTransaction(contract())
	tx.Sign(privKey)

	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return nil
	}

	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

	return msg.Bytes()
}
