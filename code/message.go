package code

import (
	"encoding/binary"
	"fmt"
	"io"
)

type messageID uint8

const (
	MChoke         messageID = 0
	MUnchoke       messageID = 1
	MInterested    messageID = 2
	MNotInterested messageID = 3
	MHave          messageID = 4
	MBitfield      messageID = 5
	MRequest       messageID = 6
	MPiece         messageID = 7
	MCancel        messageID = 8
)

type Message struct {
	ID      messageID
	Payload []byte
}

func (m *Message) Serialize() []byte {
	if m == nil {
		// Keep alive message
		return make([]byte, 4)
	}
	length := uint32(len(m.Payload) + 1)
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)
	return buf
}

func ReadMessage(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)
	if length == 0 {
		return nil, nil
	}
	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}
	message := Message{
		ID:      messageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}
	return &message, nil
}

func GetHavePayload(index int) []byte {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))
	return payload
}

func ParseHavePayload(message Message) (int, error) {
	if message.ID != MHave {
		return 0, fmt.Errorf("Expected have message, got %d", message.ID)
	}
	if len(message.Payload) != 4 {
		return 0, fmt.Errorf("Have message has wrong length")
	}
	index := int(binary.BigEndian.Uint32(message.Payload))
	return index, nil
}

func GetRequestPayload(index, begin, length int) []byte {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload, uint32(index))
	binary.BigEndian.PutUint32(payload[4:], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:], uint32(length))
	return payload
}

func ParsePiecePayload(index int, buf []byte, message Message) (int, error) {
	if message.ID != MPiece {
		return 0, fmt.Errorf("Expected request message, got %d", message.ID)
	}

	if len(message.Payload) < 8 {
		return 0, fmt.Errorf("Request message has wrong length")
	}

	gotIndex := int(binary.BigEndian.Uint32(message.Payload[0:4]))
	begin := int(binary.BigEndian.Uint32(message.Payload[4:8]))
	block := message.Payload[8:]

	if gotIndex != index {
		return 0, fmt.Errorf("Expected index %d, got %d", index, gotIndex)
	}

	if begin > len(buf) {
		return 0, fmt.Errorf("Begin offset %d is out of bounds", begin)
	}

	if begin+len(block) > len(buf) {
		return 0, fmt.Errorf("Block is too large")
	}

	copy(buf[begin:], block)
	return len(block), nil
}
