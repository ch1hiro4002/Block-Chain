package main

import (
	"bytes"
	"encoding/binary"
	"net"
	"time"

	"github.com/ch1hiro4002/Block-Chain/core"
	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/ch1hiro4002/Block-Chain/network"
	"github.com/sirupsen/logrus"
)

func main() {
	privKey := crypto.GeneratePrivateKey()
	localNode := makeServer("LOCAL", &privKey, []string{}, ":3000", ":8888")
	go localNode.Strat()

	// remoteNode := makeServer("REMOTE", nil, []string{"127.0.0.1:3000"}, ":4000", ":9999")
	// go remoteNode.Strat()

	time.Sleep(2 * time.Second)
	txSender()
	txSender()

	select {}
}

func makeServer(id string, privKey *crypto.PrivateKey, seedNodes []string, listenAddr string, apiListenAddr string) *network.Server {
	opts := network.ServerOpts{
		ID:            id,
		PrivateKey:    privKey,
		TCPTransport:  network.NewTCPTransport(listenAddr),
		SeedNodes:     seedNodes,
		ListenAddr:    listenAddr,
		APIListenAddr: apiListenAddr,
	}

	s, err := network.NewServer(opts)
	if err != nil {
		logrus.Error(err)
	}

	return s
}

func txSender() {
	conn, err := net.Dial("tcp", "127.0.0.1:3000")
	if err != nil {
		panic(err)
	}

	privKey := crypto.GeneratePrivateKey()

	contractData := []byte{0x03, 0x0a, 0x02, 0x0a, 0x0c, 0x46, 0x0b, 0x4f, 0x0b, 0x4f, 0x0b, 0x03, 0x0a, 0x10, 0x11}

	tx := core.NewTransaction(contractData)
	tx.Sign(privKey)

	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		logrus.Error(err)
	}

	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

	frame := make([]byte, 4+len(msg.Bytes()))
	binary.BigEndian.PutUint32(frame[:4], uint32(len(msg.Bytes())))
	copy(frame[4:], msg.Bytes())
	_, err = conn.Write(frame)
	if err != nil {
		panic(err)
	}
}
