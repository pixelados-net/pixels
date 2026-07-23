// Package tags contains the GET_USER_TAGS outbound packet.
package tags

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GET_USER_TAGS.
const Header uint16 = 1255

// Definition describes the fixed GET_USER_TAGS fields.
var Definition = codec.Definition{codec.Named("roomUnitId", codec.Int32Field), codec.Named("tagCount", codec.Int32Field)}

// TagDefinition describes one public tag.
var TagDefinition = codec.Definition{codec.Named("tag", codec.StringField)}

// Encode creates a GET_USER_TAGS packet.
func Encode(roomUnitID int32, tags []string) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, Definition, codec.Int32(roomUnitID), codec.Int32(int32(len(tags))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, tag := range tags {
		payload, err = codec.AppendPayload(payload, TagDefinition, codec.String(tag))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
