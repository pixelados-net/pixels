// Package result encodes PET_BREEDING_RESULT.
package result

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_BREEDING_RESULT.
const Header uint16 = 1553

// Result stores one breeding reward view.
type Result struct {
	// StuffID identifies the granted item or pet.
	StuffID int64
	// ClassID identifies the renderer class.
	ClassID int32
	// ProductCode identifies catalog assets.
	ProductCode string
	// UserID identifies the receiving owner.
	UserID int64
	// UserName stores the receiving owner name.
	UserName string
	// RarityLevel stores the result rarity.
	RarityLevel int32
	// HasMutation reports whether the result mutated.
	HasMutation bool
}

// Encode creates PET_BREEDING_RESULT with both owner results.
func Encode(first Result, second Result) (codec.Packet, error) {
	payload, err := appendResult(nil, first)
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = appendResult(payload, second)
	return codec.Packet{Header: Header, Payload: payload}, err
}

// appendResult appends one breeding reward record.
func appendResult(dst []byte, value Result) ([]byte, error) {
	return codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.StringField, codec.Int32Field, codec.BooleanField},
		codec.Int32(int32(value.StuffID)), codec.Int32(value.ClassID), codec.String(value.ProductCode), codec.Int32(int32(value.UserID)), codec.String(value.UserName), codec.Int32(value.RarityLevel), codec.Bool(value.HasMutation))
}
