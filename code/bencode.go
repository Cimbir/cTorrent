package code

import (
	"cTorrent/models"
	"io"
	"os"

	"github.com/jackpal/bencode-go"
)

// Open parses a torrent file
func GetBTorrent(file *os.File) (*models.BTorrent, error) {
	r := io.Reader(file)
	bto := models.BTorrent{}
	err := bencode.Unmarshal(r, &bto)
	if err != nil {
		return nil, err
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	bto.Original, err = bencode.Decode(r)
	if err != nil {
		return nil, err
	}
	return &bto, nil
}
