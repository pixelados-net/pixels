// Package notification contains the CLUB_GIFT_NOTIFICATION outbound packet.
package notification

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies CLUB_GIFT_NOTIFICATION.
	Header uint16 = 2188
)

// Encode creates a CLUB_GIFT_NOTIFICATION packet.
func Encode(count int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(count))
}
