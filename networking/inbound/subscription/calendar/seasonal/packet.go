// Package seasonal contains the GET_SEASONAL_CALENDAR_DAILY_OFFER inbound packet.
package seasonal

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the GET_SEASONAL_CALENDAR_DAILY_OFFER packet identifier.
	Header uint16 = 3257
)

// Decode validates a GET_SEASONAL_CALENDAR_DAILY_OFFER packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
