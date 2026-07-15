// Package cfhchatlog contains the moderation cfhchatlog inbound packet.
package cfhchatlog

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation cfhchatlog packet.
const Header uint16 = 211

// Payload contains decoded moderation cfhchatlog fields.
type Payload struct {
	// IssueID stores the decoded wire field.
	IssueID int32
}

// Definition describes moderation cfhchatlog fields.
var Definition = codec.Definition{
	codec.Named("issueID", codec.Int32Field),
}

// Decode validates and decodes the moderation cfhchatlog packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		IssueID: values[0].Int32,
	}, nil
}
