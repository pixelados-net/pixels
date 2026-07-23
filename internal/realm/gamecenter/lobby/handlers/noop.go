package handlers

import (
	"github.com/niflaot/pixels/networking/codec"
	inaccept "github.com/niflaot/pixels/networking/inbound/gamecenter/invite/accept"
	inchat "github.com/niflaot/pixels/networking/inbound/gamecenter/session/chat"
	inexit "github.com/niflaot/pixels/networking/inbound/gamecenter/session/exit"
	infull "github.com/niflaot/pixels/networking/inbound/gamecenter/session/fullstatus"
	inagain "github.com/niflaot/pixels/networking/inbound/gamecenter/session/playagain"
	inunloaded "github.com/niflaot/pixels/networking/inbound/gamecenter/session/unloaded"
	inready "github.com/niflaot/pixels/networking/inbound/gamecenter/stage/ready"
)

// decodeNoop validates unsupported external-arena messages without inventing state.
func decodeNoop(packet codec.Packet) error {
	switch packet.Header {
	case inaccept.Header:
		_, err := inaccept.Decode(packet)
		return err
	case inchat.Header:
		_, err := inchat.Decode(packet)
		return err
	case inexit.Header:
		_, err := inexit.Decode(packet)
		return err
	case infull.Header:
		_, err := infull.Decode(packet)
		return err
	case inagain.Header:
		return inagain.Decode(packet)
	case inunloaded.Header:
		_, err := inunloaded.Decode(packet)
		return err
	case inready.Header:
		_, err := inready.Decode(packet)
		return err
	default:
		return codec.ErrUnexpectedHeader
	}
}
