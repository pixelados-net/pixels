// Package modify contains the ROOM_FILTER_WORDS_MODIFY inbound packet.
package modify

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_FILTER_WORDS_MODIFY.
	Header uint16 = 3001
)

// Payload contains unpacked word filter mutation fields.
type Payload struct {
	// RoomID identifies the room.
	RoomID int32
	// Add reports whether the word is added instead of removed.
	Add bool
	// Word stores the filter word.
	Word string
}

// Definition describes ROOM_FILTER_WORDS_MODIFY fields.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field), codec.Named("add", codec.BooleanField), codec.Named("word", codec.StringField)}

// Decode unpacks a ROOM_FILTER_WORDS_MODIFY packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{RoomID: values[0].Int32, Add: values[1].Boolean, Word: values[2].String}, nil
}
