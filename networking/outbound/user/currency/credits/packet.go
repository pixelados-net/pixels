// Package credits contains the USER_CREDITS outbound packet.
package credits

import (
	"strconv"

	"github.com/niflaot/pixels/networking/codec"
)

const (
	// Header is the USER_CREDITS packet identifier.
	Header uint16 = 3475
)

// Definition describes the USER_CREDITS payload fields.
var Definition = codec.Definition{codec.Named("credits", codec.StringField)}

// Encode creates a USER_CREDITS packet.
func Encode(amount int64) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(strconv.FormatInt(amount, 10)+".0"))
}
