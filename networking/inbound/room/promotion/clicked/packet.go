// Package clicked decodes ROOM_AD_EVENT_TAB_CLICKED telemetry.
package clicked

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies ROOM_AD_EVENT_TAB_CLICKED.
const Header uint16 = 2412

// Payload contains renderer telemetry fields.
type Payload struct {
	RoomID   int32
	RoomName string
	Source   int32
}

// Definition describes the renderer's actual three-field telemetry shape.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field), codec.Named("roomName", codec.StringField), codec.Named("source", codec.Int32Field)}

// Decode returns one tab-click telemetry record.
func Decode(packet codec.Packet) (Payload, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Payload{}, err
	}
	v, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{RoomID: v[0].Int32, RoomName: v[1].String, Source: v[2].Int32}, nil
}
