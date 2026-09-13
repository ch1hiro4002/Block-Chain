package main

import (
	"bytes"
	"encoding/binary"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ch1hiro4002/Block-Chain/core"
	"github.com/ch1hiro4002/Block-Chain/crypto"
	"github.com/ch1hiro4002/Block-Chain/network"
	"github.com/sirupsen/logrus"
)

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
			return v
	}
	return def
}

func main() {
	nodeID := envOr("NODE_ID", "NODE")
	listenAddr := envOr("LISTEN_ADDR", ":3000")
	apiListenAddr := envOr("API_LISTEN_ADDR", "")
	blockTime, _ := time.ParseDuration(envOr("BLOCK_TIME", "2s"))
	startDelay, _ := time.ParseDuration(envOr("START_DELAY", "0s"))
	stake, _ := strconv.ParseUint(envOr("STAKE", "100"), 10, 64)

	var seedNodes []string
	if raw := os.Getenv("SEED_NODES"); raw != "" {
			for _, s := range strings.Split(raw, ",") {
					if s = strings.TrimSpace(s); s != "" {
							seedNodes = append(seedNodes, s)
					}
			}
	}

	privKey := crypto.GeneratePrivateKey()
	pubKey := privKey.PublicKey()

	opts := network.ServerOpts{
			ID:            nodeID,
			PrivateKey:    &privKey,
			TCPTransport:  network.NewTCPTransport(listenAddr),
			SeedNodes:     seedNodes,
			ListenAddr:    listenAddr,
			APIListenAddr: apiListenAddr,
			BlockTime:     blockTime,
	}

	srv, err := network.NewServer(opts, pubKey.Address(), pubKey, stake)
	if err != nil {
			logrus.Fatal(err)
	}

	if startDelay > 0 {
			logrus.Infof("node %s sleeping %s before start", nodeID, startDelay)
			time.Sleep(startDelay)
	}

	srv.Strat()
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
