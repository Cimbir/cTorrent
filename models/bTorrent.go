package models

type BTorrent struct {
	Announce     string `benconde:"announce"`
	Comment      string `benconde:"comment"`
	CreatedBy    string `benconde:"created by"`
	CreationDate int64  `benconde:"creation date"`
	Info         BInfo  `benconde:"info"`
	Original     interface{}
}

type BInfo struct {
	Length      int64   `benconde:"length"`
	Name        string  `benconde:"name"`
	PieceLength int     `benconde:"piece length"`
	Pieces      string  `benconde:"pieces"`
	Files       []BFile `benconde:"files"`
}

type BFile struct {
	Length int64    `benconde:"length"`
	Path   []string `benconde:"path"`
	Attr   string   `benconde:"attr"`
	SHA1   string   `benconde:"sha1"`
}
