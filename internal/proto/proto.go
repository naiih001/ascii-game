package proto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"ascii-game/internal/sim"
)

const (
	MsgJoin     byte = 0x01
	MsgJoinAck  byte = 0x02
	MsgMove     byte = 0x03
	MsgSnapshot byte = 0x04
)

type Message struct {
	Type    byte
	Payload []byte
}

func EncodeMessage(msg Message) []byte {
	buf := make([]byte, 5+len(msg.Payload))
	binary.BigEndian.PutUint32(buf[:4], uint32(1+len(msg.Payload)))
	buf[4] = msg.Type
	copy(buf[5:], msg.Payload)
	return buf
}

func DecodeMessage(r io.Reader) (Message, error) {
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return Message{}, err
	}

	if length == 0 {
		return Message{}, errors.New("invalid zero-length message")
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return Message{}, err
	}

	return Message{
		Type:    buf[0],
		Payload: buf[1:],
	}, nil
}

func DecodeMessageBytes(buf []byte) (Message, error) {
	return DecodeMessage(bytes.NewReader(buf))
}

func MarshalJoin() []byte {
	return nil
}

func UnmarshalJoin(payload []byte) error {
	if len(payload) != 0 {
		return fmt.Errorf("join payload must be empty, got %d bytes", len(payload))
	}

	return nil
}

func MarshalJoinAck(playerID sim.PlayerID) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(playerID))
	return buf
}

func UnmarshalJoinAck(payload []byte) (sim.PlayerID, error) {
	if len(payload) != 8 {
		return 0, fmt.Errorf("join ack payload must be 8 bytes, got %d", len(payload))
	}

	return sim.PlayerID(binary.BigEndian.Uint64(payload)), nil
}

func MarshalMove(dx, dy int) ([]byte, error) {
	if dx < -128 || dx > 127 || dy < -128 || dy > 127 {
		return nil, fmt.Errorf("move delta out of range: %d,%d", dx, dy)
	}

	return []byte{byte(int8(dx)), byte(int8(dy))}, nil
}

func UnmarshalMove(payload []byte) (int, int, error) {
	if len(payload) != 2 {
		return 0, 0, fmt.Errorf("move payload must be 2 bytes, got %d", len(payload))
	}

	return int(int8(payload[0])), int(int8(payload[1])), nil
}

func MarshalSnapshot(snapshot sim.Snapshot) []byte {
	playerCount := len(snapshot.Players)
	buf := make([]byte, 8+2+(playerCount*20))
	binary.BigEndian.PutUint64(buf[0:8], uint64(snapshot.Tick))
	binary.BigEndian.PutUint16(buf[8:10], uint16(playerCount))

	offset := 10
	for _, player := range snapshot.Players {
		binary.BigEndian.PutUint64(buf[offset:offset+8], uint64(player.ID))
		offset += 8
		binary.BigEndian.PutUint32(buf[offset:offset+4], uint32(int32(player.Pos.X)))
		offset += 4
		binary.BigEndian.PutUint32(buf[offset:offset+4], uint32(int32(player.Pos.Y)))
		offset += 4
		binary.BigEndian.PutUint32(buf[offset:offset+4], uint32(player.Sprite))
		offset += 4
	}

	return buf
}

func UnmarshalSnapshot(payload []byte) (sim.Snapshot, error) {
	if len(payload) < 10 {
		return sim.Snapshot{}, fmt.Errorf("snapshot payload too short: %d", len(payload))
	}

	tick := int64(binary.BigEndian.Uint64(payload[0:8]))
	playerCount := int(binary.BigEndian.Uint16(payload[8:10]))
	expected := 10 + (playerCount * 20)
	if len(payload) != expected {
		return sim.Snapshot{}, fmt.Errorf("snapshot payload size mismatch: got %d want %d", len(payload), expected)
	}

	players := make([]sim.PlayerState, 0, playerCount)
	offset := 10
	for i := 0; i < playerCount; i++ {
		player := sim.PlayerState{
			ID: sim.PlayerID(binary.BigEndian.Uint64(payload[offset : offset+8])),
		}
		offset += 8
		player.Pos.X = int(int32(binary.BigEndian.Uint32(payload[offset : offset+4])))
		offset += 4
		player.Pos.Y = int(int32(binary.BigEndian.Uint32(payload[offset : offset+4])))
		offset += 4
		player.Sprite = rune(binary.BigEndian.Uint32(payload[offset : offset+4]))
		offset += 4
		players = append(players, player)
	}

	return sim.Snapshot{
		Tick:    tick,
		Players: players,
	}, nil
}
