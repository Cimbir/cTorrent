package code

import (
	"bytes"
	"cTorrent/models"
	"crypto/sha1"
	"fmt"
)

type PieceTask struct {
	Index  int
	Hash   [20]byte
	Length int
}

type PieceResult struct {
	Index int
	buf   []byte
}

type PieceProgress struct {
	task *PieceTask
	Conn *PeerCommunication
	Buf  []byte

	ReqestedByteAmount int
	ReceivedByteAmount int
	Backlog            int
}

const MAXBACKLOG = 5

const MAXBLOCKSIZE = 16384

func SendRequests(progress *PieceProgress) error {
	if !progress.Conn.Choked {
		for progress.Backlog < MAXBACKLOG && progress.ReqestedByteAmount < progress.task.Length {

			requestAmount := MAXBLOCKSIZE
			if requestAmount+progress.ReqestedByteAmount > progress.task.Length {
				requestAmount = progress.task.Length - progress.ReqestedByteAmount
			}

			err := progress.Conn.SendRequest(progress.task.Index, progress.ReqestedByteAmount, requestAmount)
			if err != nil {
				return err
			}

			progress.Backlog++
			progress.ReqestedByteAmount += requestAmount

		}
	}

	return nil
}

func HandleMessages(progress *PieceProgress) error {
	msg, err := progress.Conn.ReadMessage()
	if err != nil {
		return err
	}

	if msg == nil {
		return nil
	}

	switch msg.ID {
	case MChoke:
		progress.Conn.Choked = true
	case MUnchoke:
		progress.Conn.Choked = false
	case MHave:
		index, err := ParseHavePayload(*msg)
		if err != nil {
			return err
		}
		progress.Conn.Bitfield.SetPiece(index)
	case MPiece:
		n, err := ParsePiecePayload(progress.task.Index, progress.Buf, *msg)
		if err != nil {
			return err
		}

		progress.ReceivedByteAmount += n
		progress.Backlog--
	}

	return nil
}

func DownloadPiece(peerCom *PeerCommunication, task *PieceTask) ([]byte, error) {
	progress := PieceProgress{
		task: task,
		Conn: peerCom,
		Buf:  make([]byte, task.Length),

		ReceivedByteAmount: 0,
		ReqestedByteAmount: 0,
		Backlog:            0,
	}

	for progress.ReceivedByteAmount < task.Length {
		err := SendRequests(&progress)
		if err != nil {
			return nil, err
		}

		err = HandleMessages(&progress)
		if err != nil {
			return nil, err
		}
	}

	return progress.Buf, nil
}

func IsPieceWhole(buf []byte, hash [20]byte) bool {
	byteHash := sha1.Sum(buf)
	return bytes.Equal(byteHash[:], hash[:])
}

func StartDownloadWorker(t *models.Torrent, peer Peer, taskQueue chan *PieceTask, resultQueue chan *PieceResult) {
	fmt.Println("Connecting to", peer.String())
	peerId := [20]byte{}
	copy(peerId[:], "-UT0001-123456789012")

	peerCom, err := GetPeerConnection(peer, t.InfoHash, peerId)
	if err != nil {
		fmt.Println("Failed to connect to", peer.String())
		fmt.Println(err)
		return
	}
	defer peerCom.Conn.Close()

	peerCom.SendUnchoke()
	peerCom.SendInterested()

	for task := range taskQueue {

		if !peerCom.Bitfield.HasPiece(task.Index) {
			fmt.Println("Peer", peer.String(), "does not have piece", task.Index)
			taskQueue <- task
			continue
		}

		buf, err := DownloadPiece(&peerCom, task)
		if err != nil {
			fmt.Println("Failed to download piece", task.Index, "from", peer.String())
			fmt.Println(err)
			taskQueue <- task
			continue
		}

		if !IsPieceWhole(buf, task.Hash) {
			fmt.Println("Piece", task.Index, "from", peer.String(), "is invalid")
			taskQueue <- task
			continue
		}

		peerCom.SendHave(uint32(task.Index))

		resultQueue <- &PieceResult{
			Index: task.Index,
			buf:   buf,
		}

	}
}

func GetPieceEnds(t *models.Torrent, index int) (begin int, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > int(t.Length) {
		end = int(t.Length)
	}
	return begin, end
}

func GetPieceLength(t *models.Torrent, index int) int {
	begin, end := GetPieceEnds(t, index)
	return end - begin
}

func StartDownload(t *models.Torrent, peers []Peer) ([]byte, error) {
	taskQueue := make(chan *PieceTask, len(t.PieceHashes))
	resultQueue := make(chan *PieceResult)
	for i, hash := range t.PieceHashes {
		taskQueue <- &PieceTask{
			Index:  i,
			Hash:   hash,
			Length: GetPieceLength(t, i),
		}
	}
	fmt.Println("Starting download with", len(peers), "peers")
	fmt.Println("Downloading", len(t.PieceHashes), "pieces")

	for _, peer := range peers {
		go StartDownloadWorker(t, peer, taskQueue, resultQueue)
	}

	buf := make([]byte, t.Length)
	done := 0
	for done < len(t.PieceHashes) {
		res := <-resultQueue
		fmt.Println("Got piece", res.Index)
		done++
		begin, end := GetPieceEnds(t, res.Index)
		copy(buf[begin:end], res.buf)
	}
	close(taskQueue)

	return buf, nil
}
