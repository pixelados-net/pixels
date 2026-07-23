// Package searchdelete contains the NAVIGATOR_DELETE_SAVED_SEARCH inbound packet.
package searchdelete

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the NAVIGATOR_DELETE_SAVED_SEARCH packet identifier.
	Header uint16 = 1954
)

// Payload contains the unpacked NAVIGATOR_DELETE_SAVED_SEARCH fields.
type Payload struct {
	// SearchID identifies the saved search to delete.
	SearchID int32
}

// Definition describes the NAVIGATOR_DELETE_SAVED_SEARCH payload fields.
var Definition = codec.Definition{codec.Named("searchId", codec.Int32Field)}

// Decode unpacks a NAVIGATOR_DELETE_SAVED_SEARCH packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{SearchID: values[0].Int32}, nil
}
