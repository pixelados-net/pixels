// Package postitopen opens Nitro's initial post-it editor.
package postitopen

import "github.com/niflaot/pixels/networking/codec"

// Header is the FURNITURE_POSTIT_STICKY_POLE_OPEN identifier.
const Header uint16 = 2366

// Encode creates the editor-open packet.
func Encode(itemID int64, wallPosition string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(int32(itemID)), codec.String(wallPosition))
}
