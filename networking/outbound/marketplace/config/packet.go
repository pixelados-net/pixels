// Package config contains the MARKETPLACE_CONFIG outbound packet.
package config

import (
	marketcore "github.com/niflaot/pixels/internal/realm/marketplace/core"
	"github.com/niflaot/pixels/networking/codec"
	"time"
)

// Header identifies MARKETPLACE_CONFIG.
const Header uint16 = 1823

// Encode creates MARKETPLACE_CONFIG.
func Encode(value marketcore.Options) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.BooleanField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Bool(value.Enabled), codec.Int32(int32(value.CommissionPercent)), codec.Int32(int32(value.TokenCost)), codec.Int32(int32(value.AdvertisementCost)), codec.Int32(int32(value.MinimumPrice)), codec.Int32(int32(value.MaximumPrice)), codec.Int32(int32(value.OfferDuration/time.Hour)), codec.Int32(int32(value.DisplayDuration/time.Hour)))
	return codec.Packet{Header: Header, Payload: payload}, err
}
