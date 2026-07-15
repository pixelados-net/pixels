// Package tool contains MODERATION_TOOL initialization.
package tool

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/moderation/staff/issueinfo"
)

// Header identifies MODERATION_TOOL.
const Header uint16 = 2696

// Permissions stores Nitro's seven moderator feature flags.
type Permissions struct {
	CFH       bool
	Chatlogs  bool
	Alert     bool
	Kick      bool
	Ban       bool
	RoomAlert bool
	RoomKick  bool
}

// Encode creates moderator tool initialization.
func Encode(issues []issueinfo.Params, presets []string, roomPresets []string, permissions Permissions) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(issues))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, issue := range issues {
		packet, encodeErr := issueinfo.Encode(issue)
		if encodeErr != nil {
			return codec.Packet{}, encodeErr
		}
		payload = append(payload, packet.Payload...)
	}
	payload, err = appendStrings(payload, presets)
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = appendStrings(payload, nil)
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.BooleanField, codec.BooleanField, codec.BooleanField, codec.BooleanField, codec.BooleanField, codec.BooleanField, codec.BooleanField}, codec.Bool(permissions.CFH), codec.Bool(permissions.Chatlogs), codec.Bool(permissions.Alert), codec.Bool(permissions.Kick), codec.Bool(permissions.Ban), codec.Bool(permissions.RoomAlert), codec.Bool(permissions.RoomKick))
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = appendStrings(payload, roomPresets)
	return codec.Packet{Header: Header, Payload: payload}, err
}

// appendStrings appends one int32-counted string list.
func appendStrings(payload []byte, values []string) ([]byte, error) {
	next, err := codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(values))))
	if err != nil {
		return payload, err
	}
	for _, value := range values {
		next, err = codec.AppendPayload(next, codec.Definition{codec.StringField}, codec.String(value))
		if err != nil {
			return payload, err
		}
	}
	return next, nil
}
