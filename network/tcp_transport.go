package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

const maxRPCFrameSize = 64 << 20 // 64MB

type TCPPeer struct {
	sendMu sync.Mutex
	conn   net.Conn
}

// Send sends a length-prefixed message over the TCP connection.
func (p *TCPPeer) Send(b []byte) error {
	p.sendMu.Lock()
	defer p.sendMu.Unlock()

	frame := make([]byte, 4+len(b))
	binary.BigEndian.PutUint32(frame[:4], uint32(len(b)))
	copy(frame[4:], b)
	_, err := p.conn.Write(frame)
	return err
}

// readLoop continuously reads and dispatches RPC messages from the connection.
func (p *TCPPeer) readLoop(rpcCh chan RPC) {
	for {
		header := make([]byte, 4)
		if _, err := io.ReadFull(p.conn, header); err != nil {
			fmt.Printf("read error: %v", err)
			return
		}

		size := binary.BigEndian.Uint32(header)
		if size == 0 || size > maxRPCFrameSize {
			fmt.Printf("invalid rpc frame size: %d\n", size)
			return
		}

		payload := make([]byte, size)

		if _, err := io.ReadFull(p.conn, payload); err != nil {
			fmt.Printf("read error: %v", err)
			return
		}

		rpcCh <- RPC{
			From:    NetAddr(p.conn.RemoteAddr().String()),
			Payload: bytes.NewReader(payload),
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
