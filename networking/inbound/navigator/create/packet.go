// Package create contains the ROOM_CREATE inbound packet.
package create

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_CREATE packet identifier.
	Header uint16 = 2752
)

// Payload contains the unpacked ROOM_CREATE fields.
type Payload struct {
	// RoomName stores the requested room name.
	RoomName string
	// RoomDescription stores the requested room description.
	RoomDescription string
	// ModelName stores the requested layout model.
	ModelName string
	// CategoryID identifies the requested category.
	CategoryID int32
	// MaxVisitors stores the requested visitor limit.
	MaxVisitors int32
	// TradeType stores the requested trade setting.
	TradeType int32
}

// Definition describes the ROOM_CREATE payload fields.
var Definition = codec.Definition{
	codec.Named("roomName", codec.StringField),
	codec.Named("roomDescription", codec.StringField),
	codec.Named("modelName", codec.StringField),
	codec.Named("categoryId", codec.Int32Field),
	codec.Named("maxVisitors", codec.Int32Field),
	codec.Named("tradeType", codec.Int32Field),
}

// Decode unpacks a ROOM_CREATE packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}

	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return payloadFromValues(values), nil
}

// payloadFromValues returns typed ROOM_CREATE payload data.
func payloadFromValues(values []codec.Value) Payload {
	return Payload{
		RoomName:        values[0].String,
		RoomDescription: values[1].String,
		ModelName:       values[2].String,
		CategoryID:      values[3].Int32,
		MaxVisitors:     values[4].Int32,
		TradeType:       values[5].Int32,
	}
}
