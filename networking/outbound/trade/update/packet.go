// Package update contains the TRADE_UPDATE outbound packet.
package update

import (
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies TRADE_UPDATE.
const Header uint16 = 2024

// Item joins one offered item and definition.
type Item struct {
	// Item stores the durable furniture instance.
	Item furnituremodel.Item
	// Definition stores protocol furniture metadata.
	Definition furnituremodel.Definition
}

// Participant stores one participant offer.
type Participant struct {
	// PlayerID identifies the durable participant account.
	PlayerID int64
	// Items stores the complete offered furniture projection.
	Items []Item
	// RedeemableCredits stores offered credit-furniture value.
	RedeemableCredits int64
}

// Encode creates TRADE_UPDATE.
func Encode(first Participant, second Participant) (codec.Packet, error) {
	payload, err := appendParticipant(nil, first)
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = appendParticipant(payload, second)
	return codec.Packet{Header: Header, Payload: payload}, err
}

// appendParticipant appends one trade participant.
func appendParticipant(dst []byte, participant Participant) ([]byte, error) {
	dst, err := codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(participant.PlayerID)), codec.Int32(int32(len(participant.Items))))
	if err != nil {
		return nil, err
	}
	for _, entry := range participant.Items {
		typeCode := "S"
		if entry.Definition.Kind == "wall" {
			typeCode = "I"
		}
		dst, err = codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.BooleanField}, codec.Int32(int32(entry.Item.ID)), codec.String(typeCode), codec.Int32(int32(entry.Item.ID)), codec.Int32(int32(entry.Definition.SpriteID)), codec.Int32(0), codec.Bool(entry.Definition.AllowInventoryStack && entry.Item.LimitedEditionNumber == nil))
		if err != nil {
			return nil, err
		}
		if entry.Item.LimitedEditionNumber == nil {
			dst, err = codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(0), codec.String(entry.Item.ExtraData))
		} else {
			dst, err = codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(1), codec.Int32(*entry.Item.LimitedEditionNumber), codec.Int32(0))
		}
		if err != nil {
			return nil, err
		}
		dst, err = codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(0), codec.Int32(0), codec.Int32(0))
		if err == nil && typeCode == "S" {
			dst, err = codec.AppendPayload(dst, codec.Definition{codec.Int32Field}, codec.Int32(0))
		}
		if err != nil {
			return nil, err
		}
	}
	return codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(len(participant.Items))), codec.Int32(int32(participant.RedeemableCredits)))
}
