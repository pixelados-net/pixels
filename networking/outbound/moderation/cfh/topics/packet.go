// Package topics contains CFH_TOPICS projection.
package topics

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CFH_TOPICS.
const Header uint16 = 325

// Topic stores one selectable topic.
type Topic struct {
	Name        string
	ID          int32
	Consequence string
}

// Category groups topics under one localized name.
type Category struct {
	Name   string
	Topics []Topic
}

// Encode creates the nested topic tree expected by Nitro.
func Encode(categories []Category) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(categories))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, category := range categories {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.Int32Field}, codec.String(category.Name), codec.Int32(int32(len(category.Topics))))
		if err != nil {
			return codec.Packet{}, err
		}
		for _, topic := range category.Topics {
			payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.Int32Field, codec.StringField}, codec.String(topic.Name), codec.Int32(topic.ID), codec.String(topic.Consequence))
			if err != nil {
				return codec.Packet{}, err
			}
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
