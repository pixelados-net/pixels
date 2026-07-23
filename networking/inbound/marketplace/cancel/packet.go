// Package cancel contains the CANCEL_MARKETPLACE_OFFER inbound packet.
package cancel

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CANCEL_MARKETPLACE_OFFER.
const Header uint16 = 434

// Decode reads the listing id.
func Decode(packet codec.Packet) (int32, error) {
	if packet.Header != Header {
		return 0, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Named("offerId", codec.Int32Field)})
	if err != nil {
		return 0, err
	}
	return values[0].Int32, nil
}
