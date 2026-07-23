// Package training encodes PET_TRAINING_PANEL.
package training

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_TRAINING_PANEL.
const Header uint16 = 1164

// Encode creates PET_TRAINING_PANEL.
func Encode(petID int64, commands []int32, enabled []int32) (codec.Packet, error) {
	payload, err := appendList(nil, int32(petID), commands)
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = appendList(payload, int32(len(enabled)), enabled)
	return codec.Packet{Header: Header, Payload: payload}, err
}

// appendList appends a leading value, a count when needed, and integer values.
func appendList(dst []byte, leading int32, values []int32) ([]byte, error) {
	definition := codec.Definition{codec.Int32Field}
	args := []codec.Value{codec.Int32(leading)}
	if dst == nil {
		definition = append(definition, codec.Int32Field)
		args = append(args, codec.Int32(int32(len(values))))
	}
	payload, err := codec.AppendPayload(dst, definition, args...)
	if err != nil {
		return nil, err
	}
	for _, value := range values {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(value))
		if err != nil {
			return nil, err
		}
	}
	return payload, nil
}
