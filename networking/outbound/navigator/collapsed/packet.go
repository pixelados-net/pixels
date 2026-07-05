// Package collapsed contains the NAVIGATOR_COLLAPSED outbound packet.
package collapsed

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the NAVIGATOR_COLLAPSED packet identifier.
	Header uint16 = 1543
)

// Definition describes the NAVIGATOR_COLLAPSED payload fields.
var Definition = codec.Definition{codec.Named("categoryCount", codec.Int32Field)}

// CategoryDefinition describes one collapsed category code.
var CategoryDefinition = codec.Definition{codec.Named("category", codec.StringField)}

// Encode creates a NAVIGATOR_COLLAPSED packet.
func Encode(categories []string) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, Definition, codec.Int32(int32(len(categories))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, category := range categories {
		payload, err = codec.AppendPayload(payload, CategoryDefinition, codec.String(category))
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}
