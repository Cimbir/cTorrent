package main

import (
	"cTorrent/code"
	"crypto/rand"
	"fmt"
	"os"
)

func main() {
	file, err := os.Open("../dracula.torrent")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	parsedTorrent, err := code.GetBTorrent(file)
	if err != nil {
		fmt.Println("Error parsing torrent:", err)
		return
	}

	torrent, err := parsedTorrent.ToTorrentFile()
	if err != nil {
		fmt.Println("Error converting torrent:", err)
		return
	}

	var name [20]byte
	_, err = rand.Read(name[:])
	if err != nil {
		fmt.Println("Error generating peer ID:", err)
		return
	}

	resp, err := code.GetTrackerResponse(&torrent, name, 6881)
	if err != nil {
		fmt.Println("Error getting tracker response:", err)
		return
	}

	fmt.Println("Tracker response:", resp.Interval)

	peers, err := code.GetPeers([]byte(resp.Peers))
	if err != nil {
		fmt.Println("Error getting peers:", err)
		return
	}

	fmt.Println("Peers:")
	for _, peer := range peers {

		conn, err := code.StartTCP(peer)
		if err != nil {
			fmt.Println("Error connecting to peer:", err)
			continue
		}

		handshake := code.Handshake{
			InfoHash: torrent.InfoHash,
			PeerID:   name,
		}

		err = code.CompleteHandshake(peer, conn, &handshake)
		if err != nil {
			fmt.Println("Error completing handshake:", err)
			continue
		}

		fmt.Println("Handshake successful with", peer.String())
	}
}
