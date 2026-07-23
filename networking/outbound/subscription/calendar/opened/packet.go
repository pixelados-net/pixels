// Package opened contains the CAMPAIGN_CALENDAR_DOOR_OPENED outbound packet.
package opened

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies CAMPAIGN_CALENDAR_DOOR_OPENED.
	Header uint16 = 2551
)

// Encode creates a CAMPAIGN_CALENDAR_DOOR_OPENED packet.
func Encode(success bool, productName string, customImage string, furnitureClassName string) (codec.Packet, error) {
	definition := codec.Definition{codec.BooleanField, codec.StringField, codec.StringField, codec.StringField}
	return codec.NewPacket(Header, definition, codec.Bool(success), codec.String(productName),
		codec.String(customImage), codec.String(furnitureClassName))
}
