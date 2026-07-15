// Package roominfo contains MODTOOL_ROOM_INFO projection.
package roominfo

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MODTOOL_ROOM_INFO.
const Header uint16 = 1333

// Params stores room moderation details.
type Params struct {
	RoomID      int32
	UserCount   int32
	OwnerInRoom bool
	OwnerID     int32
	OwnerName   string
	Exists      bool
	Name        string
	Description string
	Tags        []string
}

// Encode creates room moderator information.
func Encode(params Params) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.Int32Field, codec.StringField, codec.BooleanField}, codec.Int32(params.RoomID), codec.Int32(params.UserCount), codec.Bool(params.OwnerInRoom), codec.Int32(params.OwnerID), codec.String(params.OwnerName), codec.Bool(params.Exists))
	if err != nil || !params.Exists {
		return codec.Packet{Header: Header, Payload: payload}, err
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.StringField, codec.Int32Field}, codec.String(params.Name), codec.String(params.Description), codec.Int32(int32(len(params.Tags))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, tag := range params.Tags {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField}, codec.String(tag))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
