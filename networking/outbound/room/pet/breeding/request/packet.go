// Package request encodes PET_CONFIRM_BREEDING_REQUEST.
package request

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_CONFIRM_BREEDING_REQUEST.
const Header uint16 = 634

// Parent stores one parent preview.
type Parent struct {
	// ID identifies the parent pet.
	ID int64
	// Name stores the visible name.
	Name string
	// Level stores the current level.
	Level int32
	// Figure stores the renderer figure string.
	Figure string
	// OwnerName stores the visible owner name.
	OwnerName string
}

// RarityCategory stores one weighted breed group.
type RarityCategory struct {
	// Chance stores the selection weight.
	Chance int32
	// Breeds stores eligible breed identifiers.
	Breeds []int32
}

// Encode creates PET_CONFIRM_BREEDING_REQUEST.
func Encode(nestID int64, first Parent, second Parent, categories []RarityCategory, resultPetType int32) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(nestID)))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, parent := range []Parent{first, second} {
		payload, err = appendParent(payload, parent)
		if err != nil {
			return codec.Packet{}, err
		}
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(categories))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, category := range categories {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(category.Chance), codec.Int32(int32(len(category.Breeds))))
		if err != nil {
			return codec.Packet{}, err
		}
		for _, breed := range category.Breeds {
			payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(breed))
			if err != nil {
				return codec.Packet{}, err
			}
		}
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(resultPetType))
	return codec.Packet{Header: Header, Payload: payload}, err
}

// appendParent appends one breeding parent preview.
func appendParent(dst []byte, value Parent) ([]byte, error) {
	return codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.StringField, codec.Int32Field, codec.StringField, codec.StringField},
		codec.Int32(int32(value.ID)), codec.String(value.Name), codec.Int32(value.Level), codec.String(value.Figure), codec.String(value.OwnerName))
}
