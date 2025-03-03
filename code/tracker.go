package code

import (
	"fmt"
	"net/http"
	"net/url"

	"cTorrent/models"

	"github.com/jackpal/bencode-go"
)

type TrackerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func GeneratePeerID() ([20]byte, error) {
	var peerID [20]byte
	copy(peerID[:], "-UT0001-123456789012")
	return peerID, nil
}

// 	copy(peerID[:], []byte("-CT0001-"))
// 	_, err := rand.Read(peerID[8:])
// 	if err != nil {
// 		return [20]byte{}, err
// 	}
// 	return peerID, nil
// }

func BuildTrackerURL(t *models.Torrent, peerID [20]byte, port uint16, announce_index int) (string, error) {
	announceURL, err := url.Parse(t.AnnounceList[announce_index])
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

func GetTrackerResponse(t *models.Torrent, peerID [20]byte, port uint16, announce_index int) (TrackerResponse, error) {
	url, err := BuildTrackerURL(t, peerID, port, announce_index)
	if err != nil {
		fmt.Println("Error building tracker URL:", err)
		return TrackerResponse{}, err
	}

	client := &http.Client{}
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
