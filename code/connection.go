package code

import (
	"fmt"
	"net"
	"time"
)

// Start TCP connection to peer

func StartTCP(peer Peer) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 5*time.Second)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Complete handshake with peer

const Protocol = 19
const Pstr = "BitTorrent protocol"
const ReservedAm = 8

type Handshake struct {
	InfoHash [20]byte
	PeerID   [20]byte
}

func (h *Handshake) Serialize() []byte {
	buf := make([]byte, 0, 49)
	buf = append(buf, byte(Protocol))
	buf = append(buf, Pstr...)
	buf = append(buf, make([]byte, ReservedAm)...)
	buf = append(buf, h.InfoHash[:]...)
	buf = append(buf, h.PeerID[:]...)
	return buf
}

func ParseHandshake(buf []byte) (*Handshake, error) {
	if len(buf) != 68 {
		return nil, fmt.Errorf("Invalid handshake length: %d", len(buf))
	}

	if buf[0] != Protocol {
		return nil, fmt.Errorf("Invalid protocol: %d", buf[0])
	}

	if string(buf[1:20]) != Pstr {
		return nil, fmt.Errorf("Invalid protocol string: %s", buf[1:20])
	}

	h := Handshake{}
	copy(h.InfoHash[:], buf[28:48])
	copy(h.PeerID[:], buf[48:68])
	return &h, nil
}

func CompleteHandshake(peer Peer, conn net.Conn, h *Handshake) error {
	defer conn.Close()

	_, err := conn.Write(h.Serialize())
	if err != nil {
		return err
	}

	response := make([]byte, 68)
	_, err = conn.Read(response)
	if err != nil {
		return err
	}

	receivedHandshake, err := ParseHandshake(response)
	if err != nil {
		return err
	}

	if receivedHandshake.InfoHash != h.InfoHash {
		return fmt.Errorf("InfoHash mismatch")
	}

	if receivedHandshake.PeerID == h.PeerID {
		return fmt.Errorf("Connected to self")
	}

	fmt.Println("Handshake sent to", peer.String())
	return nil
}
