package code

import (
	"fmt"
	"net"
	"time"
)

// Start TCP connection to peer

func StartTCP(peer Peer) (net.Conn, error) {
	for retries := 0; retries < 5; retries++ {
		conn, err := net.Dial("tcp", peer.String())
		if err != nil {
			fmt.Println("Error connecting to peer:", err)
			time.Sleep(2 * time.Second)
			continue
		}
		return conn, nil
	}
	return nil, fmt.Errorf("failed to connect to peer")
}

// Complete handshake with peer

const Protocol = 19
const Pstr = "BitTorrent protocol"
const ReservedAm = 8

type Handshake struct {
	InfoHash [20]byte
	PeerID   [20]byte
}

type PeerCommunication struct {
	Conn       net.Conn
	Choked     bool
	Interested bool
	Bitfield   BitField
	Peer       Peer
	InfoHash   [20]byte
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
		return nil, fmt.Errorf("invalid handshake length: %d", len(buf))
	}

	if buf[0] != Protocol {
		return nil, fmt.Errorf("invalid protocol: %d", buf[0])
	}

	if string(buf[1:20]) != Pstr {
		return nil, fmt.Errorf("invalid protocol string: %s", buf[1:20])
	}

	h := Handshake{}
	copy(h.InfoHash[:], buf[28:48])
	copy(h.PeerID[:], buf[48:68])
	return &h, nil
}

func CompleteHandshake(peer Peer, conn net.Conn, h *Handshake) error {
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
		return fmt.Errorf("infoHash mismatch")
	}

	if receivedHandshake.PeerID == h.PeerID {
		return fmt.Errorf("connected to self")
	}

	fmt.Println("Handshake sent to", peer.String())
	return nil
}

func ReceiveBitField(conn net.Conn) (BitField, error) {
	message, err := ReadMessage(conn)
	if err != nil {
		return nil, err
	}

	if message.ID != MBitfield {
		return nil, fmt.Errorf("expected bitfield, received %d", message.ID)
	}

	return message.Payload, nil
}

func GetPeerConnection(peer Peer, infoHash [20]byte, peerId [20]byte) (PeerCommunication, error) {
	conn, err := StartTCP(peer)
	if err != nil {
		return PeerCommunication{}, err
	}

	handshake := Handshake{
		InfoHash: infoHash,
		PeerID:   peerId,
	}

	err = CompleteHandshake(peer, conn, &handshake)
	if err != nil {
		fmt.Println("Error completing handshake:", err)
		return PeerCommunication{}, err
	}

	bitfield, err := ReceiveBitField(conn)
	if err != nil {
		fmt.Println("Error receiving bitfield:", err)
		return PeerCommunication{}, err
	}

	return PeerCommunication{
		Conn:       conn,
		Choked:     true,
		Interested: false,
		Bitfield:   bitfield,
		Peer:       peer,
		InfoHash:   infoHash,
	}, nil
}

func (c *PeerCommunication) ReadMessage() (*Message, error) {
	return ReadMessage(c.Conn)
}

func (c *PeerCommunication) SendChoke() error {
	toSend := Message{
		ID: MChoke,
	}
	_, err := c.Conn.Write(toSend.Serialize())
	return err
}

func (c *PeerCommunication) SendUnchoke() error {
	toSend := Message{
		ID: MUnchoke,
	}
	_, err := c.Conn.Write(toSend.Serialize())
	return err
}

func (c *PeerCommunication) SendInterested() error {
	toSend := Message{
		ID: MInterested,
	}
	_, err := c.Conn.Write(toSend.Serialize())
	return err
}

func (c *PeerCommunication) SendNotInterested() error {
	toSend := Message{
		ID: MNotInterested,
	}
	_, err := c.Conn.Write(toSend.Serialize())
	return err
}

func (c *PeerCommunication) SendHave(index uint32) error {
	toSend := Message{
		ID:      MHave,
		Payload: GetHavePayload(int(index)),
	}
	_, err := c.Conn.Write(toSend.Serialize())
	return err
}

func (c *PeerCommunication) SendRequest(index, begin, len int) error {
	toSend := Message{
		ID:      MRequest,
		Payload: GetRequestPayload(index, begin, len),
	}
	_, err := c.Conn.Write(toSend.Serialize())
	return err
}
