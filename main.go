package main

import (
	"bytes"
	"net"

	"github.com/ch1hiro4002/Block-Chain/core"
	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/ch1hiro4002/Block-Chain/network"
	"github.com/sirupsen/logrus"
)

func main() {
	privKey := crypto.GeneratePrivateKey()
	localNode := makeServer("LOCAL", &privKey, ":3000", []string{"127.0.0.1:4000"})
	go localNode.Strat()

	remoeteNode := makeServer("Remote", nil, ":4000", []string{"127.0.0.1:3000"})
	go remoeteNode.Strat()

	select {}
}

func tcpTester() {
	conn, err := net.Dial("tcp", "127.0.0.1:3000")
	if err != nil {
		panic(err)
	}

	_, err = conn.Write(createTxMessage())
	if err != nil {
		panic(err)
	}
}

func makeServer(id string, privKey *crypto.PrivateKey, listenAddr string, seedNodes []string) *network.Server {
	opts := network.ServerOpts{
		ID:           id,
		PrivateKey:   privKey,
		TCPTransport: network.NewTCPTransport(listenAddr),
		ListenAddr:   listenAddr,
		SeedNodes:    seedNodes,
	}

	s, err := network.NewServer(opts)
	if err != nil {
		logrus.Error(err)
	}

	return s
}

func createTxMessage() []byte {
	privKey := crypto.GeneratePrivateKey()

	contractData := []byte{0x03, 0x0a, 0x02, 0x0a, 0x0c, 0x46, 0x0b, 0x4f, 0x0b, 0x4f, 0x0b, 0x03, 0x0a, 0x10, 0x11}

	tx := core.NewTransaction(contractData)
	tx.Sign(privKey)

	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		logrus.Error(err)
	}

	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

	return msg.Bytes()
}
