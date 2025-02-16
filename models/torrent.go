package models

import (
	"bytes"
	"crypto/sha1"
	"fmt"

	"github.com/jackpal/bencode-go"
)

type Torrent struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int64
	Name        string
	Files       []TFile
	Original    interface{}
}

type TFile struct {
	Path   []string
	Length int64
	Offset int64
}

func (bto *BTorrent) GetInfoHash() ([20]byte, error) {
	info := bto.Original.(map[string]interface{})["info"]

	var buf bytes.Buffer
	err := bencode.Marshal(&buf, info)
	if err != nil {
		return [20]byte{}, err
	}

	return sha1.Sum(buf.Bytes()), nil
}

func (i *BInfo) SplitPieceHashes() ([][20]byte, error) {
	hashLen := 20 // Length of SHA-1 hash
	buf := []byte(i.Pieces)
	if len(buf)%hashLen != 0 {
		err := fmt.Errorf("Received malformed pieces of length %d", len(buf))
		return nil, err
	}
	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}

func (bto *BTorrent) ToTorrentFile() (Torrent, error) {
	infoHash, err := bto.GetInfoHash()
	if err != nil {
		return Torrent{}, err
	}

	pieceHashes, err := bto.Info.SplitPieceHashes()
	if err != nil {
		return Torrent{}, err
	}

	res := Torrent{
		Announce:    bto.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bto.Info.PieceLength,
		Length:      0,
		Name:        bto.Info.Name,
		Original:    bto.Original,
	}

	offset := int64(0)
	for _, file := range bto.Info.Files {
		res.Files = append(res.Files, TFile{
			Path:   file.Path,
			Length: file.Length,
			Offset: offset,
		})
		offset += file.Length
		res.Length += file.Length
	}

	return res, nil
}
