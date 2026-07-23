// Package reportingstatus contains the moderation reportingstatus outbound packet.
package reportingstatus

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation reportingstatus packet.
const Header uint16 = 3463

// Definition describes moderation reportingstatus fields.
var Definition = codec.Definition{
	codec.Named("status", codec.Int32Field),
	codec.Named("reports", codec.Int32Field),
	codec.Named("limit", codec.Int32Field),
	codec.Named("enabled", codec.BooleanField),
	codec.Named("field5", codec.StringField),
	codec.Named("field6", codec.StringField),
	codec.Named("field7", codec.StringField),
	codec.Named("field8", codec.StringField),
}

// Encode creates a moderation reportingstatus packet.
func Encode(status int32, reports int32, limit int32, enabled bool, field5 string, field6 string, field7 string, field8 string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(status), codec.Int32(reports), codec.Int32(limit), codec.Bool(enabled), codec.String(field5), codec.String(field6), codec.String(field7), codec.String(field8))
}
