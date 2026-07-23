// Package common decodes shared WIRED configuration fields.
package common

import (
	"github.com/niflaot/pixels/networking/codec"
)

const (
	// MaxIntParams bounds client-controlled integer settings.
	MaxIntParams int32 = 32
	// MaxSelectedItems bounds client-controlled selected furniture identifiers.
	MaxSelectedItems int32 = 128
)

// Settings stores the common WIRED save payload.
type Settings struct {
	// ItemID identifies the configured WIRED furniture item.
	ItemID int32
	// IntParams stores descriptor-specific integer values.
	IntParams []int32
	// StringParam stores descriptor-specific text.
	StringParam string
	// ItemIDs stores selected room furniture identifiers.
	ItemIDs []int32
	// DelayPulses stores effect delay in 500 millisecond pulses.
	DelayPulses int32
	// SelectionMode stores Nitro's target selection policy.
	SelectionMode int32
}

// DecodeSettings decodes the shared variable-length save shape.
func DecodeSettings(payload []byte, withDelay bool) (Settings, error) {
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, payload)
	if err != nil {
		return Settings{}, err
	}
	settings := Settings{ItemID: values[0].Int32}
	settings.IntParams, rest, err = decodeInts(rest, values[1].Int32, MaxIntParams)
	if err != nil {
		return Settings{}, err
	}
	values, rest, err = codec.DecodePayload(nil, codec.Definition{codec.StringField, codec.Int32Field}, rest)
	if err != nil {
		return Settings{}, err
	}
	settings.StringParam = values[0].String
	settings.ItemIDs, rest, err = decodeInts(rest, values[1].Int32, MaxSelectedItems)
	if err != nil {
		return Settings{}, err
	}
	definition := codec.Definition{codec.Int32Field}
	if withDelay {
		definition = append(definition, codec.Int32Field)
	}
	values, rest, err = codec.DecodePayload(nil, definition, rest)
	if err != nil {
		return Settings{}, err
	}
	if withDelay {
		settings.DelayPulses = values[0].Int32
		settings.SelectionMode = values[1].Int32
	} else {
		settings.SelectionMode = values[0].Int32
	}
	if len(rest) != 0 {
		return Settings{}, codec.ErrUnexpectedPayload
	}
	return settings, nil
}

// decodeInts decodes one bounded integer vector.
func decodeInts(payload []byte, count int32, maximum int32) ([]int32, []byte, error) {
	if count < 0 || count > maximum {
		return nil, payload, codec.ErrUnexpectedPayload
	}
	values := make([]int32, count)
	rest := payload
	for index := range values {
		decoded, next, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field}, rest)
		if err != nil {
			return nil, payload, err
		}
		values[index] = decoded[0].Int32
		rest = next
	}
	return values, rest, nil
}
