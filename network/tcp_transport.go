package network

import (
	"bytes"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
)

type TCPPeer struct {
	conn net.Conn
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.conn.Write(b)
	return err
}

func (p *TCPPeer) readLoop(rpcCh chan RPC) {
	buf := make([]byte, 4096)

	for {
		n, err := p.conn.Read(buf)
		if err != nil {
			fmt.Printf("read error: %v", err)
			return
		}

		msg := buf[:n]

		rpcCh<-RPC {
			From: NetAddr(p.conn.RemoteAddr().String()),
			Payload: bytes.NewReader(msg),
		}
	}

}

type TCPTransport struct {
	listenAddr string
	listener   net.Listener
	peerCh     chan *TCPPeer
}

func NewTCPTransport(addr string) *TCPTransport {
	return &TCPTransport{
		listenAddr: addr,
		peerCh:     make(chan *TCPPeer),
	}
}

func (t *TCPTransport) Start() {
	listener, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		logrus.Error(err)
		return
	}

	t.listener = listener

	go t.acceptLoop()
}

func (t *TCPTransport) acceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("accept error: %v\n", err)
		}

		peer := &TCPPeer{
			conn: conn,
		}

		t.peerCh <- peer
	}
}
