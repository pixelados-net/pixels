// Package offer encodes POLL_OFFER responses.
package offer

import "github.com/niflaot/pixels/networking/codec"

// Header identifies POLL_OFFER.
const Header uint16 = 3785

// Encode creates one POLL_OFFER response.
func Encode(id int32, pollType string, headline string, summary string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.StringField, codec.StringField, codec.StringField}, codec.Int32(id), codec.String(pollType), codec.String(headline), codec.String(summary))
}
