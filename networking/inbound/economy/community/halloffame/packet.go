// Package halloffame decodes the GET_COMMUNITY_GOAL_HALL_OF_FAME request.
package halloffame

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GET_COMMUNITY_GOAL_HALL_OF_FAME.
const Header uint16 = 2167

// Definition describes the requested community goal code.
var Definition = codec.Definition{codec.Named("goalCode", codec.StringField)}

// Decode returns the requested community goal code.
func Decode(packet codec.Packet) (string, error) {
	if packet.Header != Header {
		return "", codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return "", err
	}
	return values[0].String, nil
}
