package code

import (
	"fmt"
	"net"
)

const PeerSize = 6 // 4 bytes for IP, 2 bytes for port

type Peer struct {
	IP   net.IP
	Port uint16
}

func GetPeers(peers []byte) ([]Peer, error) {
	if len(peers)%PeerSize != 0 {
		err := fmt.Errorf("Received malformed peers of length %d", len(peers))
		return nil, err
	}
	amount := len(peers) / PeerSize
	var res []Peer

	for i := 0; i < amount; i++ {
		start := i * PeerSize
		ip := net.IP(peers[start : start+4])
		port := uint16(peers[start+4])<<8 | uint16(peers[start+5])
		res = append(res, Peer{ip, port})
	}

	return res, nil
}

func (p Peer) String() string {
	return fmt.Sprintf("%s:%d", p.IP.String(), p.Port)
}
