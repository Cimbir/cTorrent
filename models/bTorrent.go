package models

type BTorrent struct {
	Announce     string     `bencode:"announce"`
	AnnounceList [][]string `bencode:"announce-list"`
	Comment      string     `bencode:"comment"`
	CreatedBy    string     `bencode:"created by"`
	CreationDate int64      `bencode:"creation date"`
	Info         BInfo      `bencode:"info"`
	Original     interface{}
}

type BInfo struct {
	Length      int64   `bencode:"length"`
	Name        string  `bencode:"name"`
	PieceLength int     `bencode:"piece length"`
	Pieces      string  `bencode:"pieces"`
	Files       []BFile `bencode:"files"`
}

type BFile struct {
	Length int64    `bencode:"length"`
	Path   []string `bencode:"path"`
	Attr   string   `bencode:"attr"`
	SHA1   string   `bencode:"sha1"`
}
