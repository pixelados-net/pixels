// Package craft contains the CRAFTING_RESULT outbound packet.
package craft

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CRAFTING_RESULT.
const Header uint16 = 618

// Option configures conditional successful result fields.
type Option func(*result)

// result stores optional successful result fields.
type result struct {
	// recipeName stores the crafted recipe name.
	recipeName string
	// itemName stores the granted furniture product name.
	itemName string
}

// WithProduct supplies the fields required by a successful result.
func WithProduct(recipeName string, itemName string) Option {
	return func(value *result) {
		value.recipeName = recipeName
		value.itemName = itemName
	}
}

// Encode creates one conditional crafting result packet.
func Encode(success bool, options ...Option) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.BooleanField}, codec.Bool(success))
	if err != nil || !success {
		return codec.Packet{Header: Header, Payload: payload}, err
	}
	value := result{}
	for _, option := range options {
		option(&value)
	}
	if value.recipeName == "" || value.itemName == "" {
		return codec.Packet{}, codec.ErrInvalidField
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.StringField}, codec.String(value.recipeName), codec.String(value.itemName))
	return codec.Packet{Header: Header, Payload: payload}, err
}
