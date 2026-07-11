// Package blockedtiles contains the ROOM_MODEL_BLOCKED_TILES outbound packet.
package blockedtiles

import (
	"encoding/binary"
	"errors"

	"github.com/niflaot/pixels/networking/codec"
)

const (
	// Header identifies ROOM_MODEL_BLOCKED_TILES.
	Header uint16 = 3990
)

var (
	// ErrTooManyTiles reports a tile count outside the int32 protocol range.
	ErrTooManyTiles = errors.New("too many blocked room tiles")
)

// Tile stores one blocked floor plan coordinate.
type Tile struct {
	// X stores the horizontal coordinate.
	X int32
	// Y stores the vertical coordinate.
	Y int32
}

// Encode creates a ROOM_MODEL_BLOCKED_TILES packet.
func Encode(tiles []Tile) (codec.Packet, error) {
	if uint64(len(tiles)) > uint64(^uint32(0)>>1) {
		return codec.Packet{}, ErrTooManyTiles
	}
	payload := make([]byte, 4, 4+len(tiles)*8)
	binary.BigEndian.PutUint32(payload, uint32(len(tiles)))
	for _, tile := range tiles {
		payload = binary.BigEndian.AppendUint32(payload, uint32(tile.X))
		payload = binary.BigEndian.AppendUint32(payload, uint32(tile.Y))
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}
