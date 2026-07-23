// Package open contains the OPEN_CAMPAIGN_CALENDAR_DOOR inbound packet.
package open

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies OPEN_CAMPAIGN_CALENDAR_DOOR.
	Header uint16 = 2257
)

// Payload contains a calendar door request.
type Payload struct {
	// CampaignName identifies the campaign.
	CampaignName string
	// DayNumber identifies the requested door.
	DayNumber int32
}

// Definition describes the packet payload.
var Definition = codec.Definition{codec.StringField, codec.Int32Field}

// Decode decodes an OPEN_CAMPAIGN_CALENDAR_DOOR packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	v, e := codec.DecodePacketExact(packet, Definition)
	if e != nil {
		return Payload{}, e
	}
	return Payload{CampaignName: v[0].String, DayNumber: v[1].Int32}, nil
}
