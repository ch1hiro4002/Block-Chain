package main

import (
	"bytes"
	"time"

	"github.com/ch1hiro4002/Block-Chain/core"
	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/ch1hiro4002/Block-Chain/network"
	"github.com/sirupsen/logrus"
)

var transports = []network.Transport{
	network.NewLocalTransport("LOCAL"),
	network.NewLocalTransport("REMOTE"),
}

func main() {
	localNode := transports[0]
	remoteNode := transports[1]

	privKey := crypto.GeneratePrivateKey()
	localServer := makeServer("LOCAL", &privKey, transports[0])

	go func() {
		remoteServer := makeServer("REMOTE", nil, transports[1])
		remoteServer.Strat()
	}()

	go func() {
		for {
			err := sendTxMessage(localNode, remoteNode)
			if err != nil {
				logrus.Error(err)
			}
			time.Sleep(2 * time.Second)
		}
	}()

	go func() {
		time.Sleep(10 * time.Second)

		remoteServer := makeServer("LATEST_REMOTE", nil, network.NewLocalTransport("LATEST_REMOTE"))
		remoteServer.Strat()
	}()

	localServer.Strat()
}

func makeServer(id string, privKey *crypto.PrivateKey, ts network.Transport) *network.Server {
	opts := network.ServerOpts{
		ID:         id,
		PrivateKey: privKey,
		Transport:  ts,
		Transports: transports,
	}

	s, err := network.NewServer(opts)
	if err != nil {
		logrus.Error(err)
	}

	return s
}

func contract() []byte {
	fooData := []byte{0x46, 0x0b, 0x4f, 0x0b, 0x4f, 0x0b, 0x03, 0x0a, 0x10, 0x12}
	data := []byte{0x03, 0x0a, 0x02, 0x0a, 0x0c, 0x46, 0x0b, 0x4f, 0x0b, 0x4f, 0x0b, 0x03, 0x0a, 0x10, 0x11}
	data = append(data, fooData...)

	return data
}

func sendTxMessage(from, to network.Transport) error {
	privKey := crypto.GeneratePrivateKey()

	tx := core.NewTransaction(contract())
	tx.Sign(privKey)

	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

	return from.SendMessage(to.Addr(), msg.Bytes())
}
