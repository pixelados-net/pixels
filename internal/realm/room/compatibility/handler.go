// Package compatibility owns explicit retired Room protocol NOOP adapters.
package compatibility

import (
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inclothing "github.com/niflaot/pixels/networking/inbound/room/compatibility/clothing"
	innetwork "github.com/niflaot/pixels/networking/inbound/room/compatibility/network"
	inpromoted "github.com/niflaot/pixels/networking/inbound/room/compatibility/promoted"
	inqueue "github.com/niflaot/pixels/networking/inbound/room/compatibility/queue"
	invote "github.com/niflaot/pixels/networking/inbound/room/compatibility/vote"
)

// Register installs legacy packets that must decode safely without domain effects.
//
// Nitro React has no caller for these composers and the reference emulators expose
// no surviving semantics. VOTE_FOR_ROOM is also retired because Nitro's live button
// sends ROOM_LIKE 3582; mutating on both would risk counting one vote twice. Keeping
// strict decoders avoids invented room-network, queue, clothing, or promotion state.
func Register(registry *netconn.HandlerRegistry) {
	if registry == nil {
		return
	}
	for _, header := range []uint16{inpromoted.Header, inclothing.Header, inqueue.Header, innetwork.Header, invote.Header} {
		_ = registry.Register(header, handle)
	}
}

// handle validates one retired packet and intentionally performs no mutation.
func handle(_ netconn.Context, packet codec.Packet) error {
	switch packet.Header {
	case inpromoted.Header:
		_, err := inpromoted.Decode(packet)
		return err
	case inclothing.Header:
		_, err := inclothing.Decode(packet)
		return err
	case inqueue.Header:
		_, err := inqueue.Decode(packet)
		return err
	case innetwork.Header:
		_, err := innetwork.Decode(packet)
		return err
	case invote.Header:
		return invote.Decode(packet)
	default:
		return codec.ErrUnexpectedHeader
	}
}
