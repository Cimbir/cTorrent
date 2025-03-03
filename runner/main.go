package main

import (
	"cTorrent/code"
	"fmt"
	"os"
)

func main() {
	file, err := os.Open("../torrents/golem.torrent")
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

	name, err := code.GeneratePeerID()
	if err != nil {
		fmt.Println("Error generating peer ID:", err)
		return
	}
	fmt.Println("Peer ID:", name)

	var peers []code.Peer
	for i := range torrent.AnnounceList {

		resp, err := code.GetTrackerResponse(&torrent, name, 6881, i)
		if err != nil {
			fmt.Println("Error getting tracker response:", err)
			return
		}

		fmt.Println("Tracker response:", resp.Interval)

		cur_peers, err := code.GetPeers([]byte(resp.Peers))
		if err != nil {
			fmt.Println("Error getting peers:", err)
			return
		}

		peers = append(peers, cur_peers...)
	}

	uniquePeers := make(map[string]code.Peer)
	for _, peer := range peers {
		uniquePeers[peer.String()] = peer
	}

	peers = make([]code.Peer, 0, len(uniquePeers))
	for _, peer := range uniquePeers {
		peers = append(peers, peer)
	}

	fmt.Println("Peers:")
	for _, peer := range peers {
		fmt.Println(peer.String())
	}

	// fmt.Println("Peers:")
	// for _, peer := range peers {

	// 	conn, err := code.StartTCP(peer)
	// 	if err != nil {
	// 		fmt.Println("Error connecting to peer:", err)
	// 		continue
	// 	}

	// 	handshake := code.Handshake{
	// 		InfoHash: torrent.InfoHash,
	// 		PeerID:   name,
	// 	}

	// 	err = code.CompleteHandshake(peer, conn, &handshake)
	// 	if err != nil {
	// 		fmt.Println("Error completing handshake:", err)
	// 		continue
	// 	}

	// 	fmt.Println("Handshake successful with", peer.String())
	// }

	info, err := code.StartDownload(&torrent, peers)
	if err != nil {
		fmt.Println("Error downloading:", err)
		return
	}

	fmt.Println("Downloaded torrent:", len(info))

	outputFile, err := os.Create("downloaded_file")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outputFile.Close()

	_, err = outputFile.Write(info)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("File saved successfully")
}
