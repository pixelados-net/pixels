// Package cancel encodes ROOM_EVENT_CANCEL responses.
package cancel

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ROOM_EVENT_CANCEL.
const Header uint16 = 3479

// Definition describes the header-only response.
var Definition = codec.Definition{}

// Encode creates one room-event cancellation response.
func Encode() (codec.Packet, error) { return codec.NewPacket(Header, Definition) }
