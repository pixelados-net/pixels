// Package place contains the PLACE_FLOOR_ITEM inbound packet.
package place

import (
	"errors"
	"strconv"
	"strings"

	"github.com/niflaot/pixels/networking/codec"
)

const (
	// Header is the PLACE_FLOOR_ITEM packet identifier.
	Header uint16 = 1258
)

// ErrMalformedPlacement reports a placement payload that does not split into four integers.
var ErrMalformedPlacement = errors.New("malformed furniture placement payload")

// Payload contains the unpacked PLACE_FLOOR_ITEM fields.
type Payload struct {
	// ItemID identifies the inventory furniture item to place.
	ItemID int32

	// X stores the destination floor tile x coordinate.
	X int32

	// Y stores the destination floor tile y coordinate.
	Y int32

	// Rotation stores the destination floor instance rotation.
	Rotation int32

	// WallPosition stores Nitro wall coordinates for a wall item.
	WallPosition string
}

// Definition describes the PLACE_FLOOR_ITEM payload fields.
var Definition = codec.Definition{codec.Named("values", codec.StringField)}

// Decode unpacks a PLACE_FLOOR_ITEM packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return parsePlacement(values[0].String)
}

// parsePlacement parses the space-separated "itemId x y rotation" placement payload.
func parsePlacement(raw string) (Payload, error) {
	parts := strings.Fields(raw)
	if len(parts) == 4 && strings.HasPrefix(parts[1], ":w=") {
		itemID, err := strconv.Atoi(parts[0])
		if err != nil {
			return Payload{}, ErrMalformedPlacement
		}
		return Payload{ItemID: int32(itemID), WallPosition: strings.Join(parts[1:], " ")}, nil
	}
	if len(parts) != 4 {
		return Payload{}, ErrMalformedPlacement
	}

	numbers := make([]int32, len(parts))
	for index, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil {
			return Payload{}, ErrMalformedPlacement
		}
		numbers[index] = int32(value)
	}

	return Payload{ItemID: numbers[0], X: numbers[1], Y: numbers[2], Rotation: numbers[3]}, nil
}
