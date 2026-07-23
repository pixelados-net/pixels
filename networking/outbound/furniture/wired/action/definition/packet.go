// Package definition encodes the WIRED_ACTION outbound packet.
package definition

import (
	"github.com/niflaot/pixels/networking/codec"
	wiredcommon "github.com/niflaot/pixels/networking/outbound/furniture/wired/common"
)

// Header is the WIRED_ACTION packet identifier.
const Header uint16 = 1434

// Encode creates a WIRED action definition packet.
func Encode(selectionEnabled bool, selectionLimit int32, selected []int64, spriteID int32, itemID int64, stringParam string, intParams []int32, selectionMode int32, actionCode int32, delayPulses int32, conflicts []int32) (codec.Packet, error) {
	payload, err := wiredcommon.AppendTriggerable(nil, selectionEnabled, selectionLimit, selected, spriteID, itemID, stringParam, intParams, selectionMode)
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(actionCode), codec.Int32(delayPulses))
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = wiredcommon.AppendInt32s(payload, conflicts)
	return codec.Packet{Header: Header, Payload: payload}, err
}
