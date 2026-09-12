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
	privKey_1 := crypto.GeneratePrivateKey()
	localNode := makeServer("LOCAL", &privKey_1, []string{}, ":3000", ":9999", 100)
	go localNode.Strat()

	privKey_2 := crypto.GeneratePrivateKey()
	remoteNode := makeServer("REMOTE", &privKey_2, []string{"127.0.0.1:3000"}, ":4000", ":9998", 1000)
	go remoteNode.Strat()

	time.Sleep(10 * time.Second)

	privKey_3 := crypto.GeneratePrivateKey()
	lateNode := makeServer("LATE", &privKey_3, []string{"127.0.0.1:3000", "127.0.0.1:4000"}, ":5000", ":9997", 10000)
	go lateNode.Strat()

	select{}
}

func makeServer(id string, privKey *crypto.PrivateKey, seedNodes []string, listenAddr string, apiListenAddr string, stake uint64) *network.Server {
	opts := network.ServerOpts{
		ID:            id,
		PrivateKey:    privKey,
		TCPTransport:  network.NewTCPTransport(listenAddr),
		SeedNodes:     seedNodes,
		ListenAddr:    listenAddr,
		APIListenAddr: apiListenAddr,
	}

	pubKey := privKey.PublicKey()
	addr := pubKey.Address()

	s, err := network.NewServer(opts, addr, pubKey, stake)
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

	tx := core.NewTransaction(0, time.Now().Unix(), time.Now().Unix()+100,contractData)
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
