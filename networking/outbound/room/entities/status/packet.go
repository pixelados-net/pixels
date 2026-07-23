// Package status contains the UNIT_STATUS outbound packet.
package status

import (
	"strings"

	"github.com/niflaot/pixels/networking/codec"
)

const (
	// Header is the UNIT_STATUS packet identifier.
	Header uint16 = 1640
)

// Action stores one unit status action.
type Action struct {
	// Key stores the action key.
	Key string

	// Value stores the optional action value.
	Value string
}

// Unit stores one unit status record.
type Unit struct {
	// RoomIndex stores the room-local unit id.
	RoomIndex int64

	// X stores the previous tile x coordinate.
	X int32

	// Y stores the previous tile y coordinate.
	Y int32

	// Z stores the previous vertical height.
	Z string

	// HeadDirection stores the head direction.
	HeadDirection int32

	// BodyDirection stores the body direction.
	BodyDirection int32

	// Actions stores current unit actions.
	Actions []Action
}

// Encode creates a UNIT_STATUS packet.
func Encode(records []Unit) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(records))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, record := range records {
		payload, err = codec.AppendPayload(payload, unitDefinition(),
			codec.Int32(int32(record.RoomIndex)),
			codec.Int32(record.X),
			codec.Int32(record.Y),
			codec.String(record.Z),
			codec.Int32(record.HeadDirection),
			codec.Int32(record.BodyDirection),
			codec.String(actionString(record.Actions)),
		)
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}

// actionString formats actions with Nitro slash delimiters.
func actionString(actions []Action) string {
	var builder strings.Builder
	builder.WriteByte('/')
	for _, action := range actions {
		if action.Key == "" {
			continue
		}
		builder.WriteString(action.Key)
		if action.Value != "" {
			builder.WriteByte(' ')
			builder.WriteString(action.Value)
		}
		builder.WriteByte('/')
	}

	return builder.String()
}

// unitDefinition returns the unit status field order.
func unitDefinition() codec.Definition {
	return codec.Definition{
		codec.Named("roomIndex", codec.Int32Field),
		codec.Named("x", codec.Int32Field),
		codec.Named("y", codec.Int32Field),
		codec.Named("z", codec.StringField),
		codec.Named("headDirection", codec.Int32Field),
		codec.Named("bodyDirection", codec.Int32Field),
		codec.Named("actions", codec.StringField),
	}
}
