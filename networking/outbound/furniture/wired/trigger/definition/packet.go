// Package definition encodes the WIRED_TRIGGER outbound packet.
package definition

import (
	"github.com/niflaot/pixels/networking/codec"
	wiredcommon "github.com/niflaot/pixels/networking/outbound/furniture/wired/common"
)

// Header is the WIRED_TRIGGER packet identifier.
const Header uint16 = 383

// Encode creates a WIRED trigger definition packet.
func Encode(selectionEnabled bool, selectionLimit int32, selected []int64, spriteID int32, itemID int64, stringParam string, intParams []int32, selectionMode int32, triggerCode int32, conflicts []int32) (codec.Packet, error) {
	payload, err := wiredcommon.AppendTriggerable(nil, selectionEnabled, selectionLimit, selected, spriteID, itemID, stringParam, intParams, selectionMode)
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(triggerCode))
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = wiredcommon.AppendInt32s(payload, conflicts)
	return codec.Packet{Header: Header, Payload: payload}, err
}
