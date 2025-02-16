package code

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"cTorrent/models"

	"github.com/jackpal/bencode-go"
)

type TrackerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func BuildTrackerURL(t *models.Torrent, peerID [20]byte, port uint16) (string, error) {
	announceURL, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{fmt.Sprintf("%d", port)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"left":       []string{fmt.Sprintf("%d", t.Length)},
		"compact":    []string{"1"},
	}

	announceURL.RawQuery = params.Encode()
	return announceURL.String(), nil
}

func GetTrackerResponse(t *models.Torrent, peerID [20]byte, port uint16) (TrackerResponse, error) {
	url, err := BuildTrackerURL(t, peerID, port)
	if err != nil {
		fmt.Println("Error building tracker URL:", err)
		return TrackerResponse{}, err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("Error contacting tracker:", err)
		return TrackerResponse{}, err
	}
	defer resp.Body.Close()

	var trackerResponse TrackerResponse
	err = bencode.Unmarshal(resp.Body, &trackerResponse)
	if err != nil {
		fmt.Println("Error decoding tracker response:", err)
		return TrackerResponse{}, err
	}

	return trackerResponse, nil
}
