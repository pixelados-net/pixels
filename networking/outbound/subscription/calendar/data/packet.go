// Package data contains the CAMPAIGN_CALENDAR_DATA outbound packet.
package data

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies CAMPAIGN_CALENDAR_DATA.
	Header uint16 = 2531
)

// Encode creates a CAMPAIGN_CALENDAR_DATA packet.
func Encode(name string, image string, currentDay int32, totalDays int32, opened []int32, missed []int32) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.StringField, codec.StringField, codec.Int32Field,
		codec.Int32Field, codec.Int32Field}, codec.String(name), codec.String(image), codec.Int32(currentDay),
		codec.Int32(totalDays), codec.Int32(int32(len(opened))))
	payload, err = appendDays(payload, opened, err)
	if err == nil {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(missed))))
	}
	payload, err = appendDays(payload, missed, err)

	return codec.Packet{Header: Header, Payload: payload}, err
}

// appendDays appends one calendar day list.
func appendDays(payload []byte, days []int32, previous error) ([]byte, error) {
	if previous != nil {
		return payload, previous
	}
	var err error
	for _, day := range days {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(day))
		if err != nil {
			return payload, err
		}
	}

	return payload, nil
}
