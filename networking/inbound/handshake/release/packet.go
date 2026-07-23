// Package release contains the RELEASE_VERSION inbound packet.
package release

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the RELEASE_VERSION packet identifier.
	Header uint16 = 4000
)

// Payload contains the unpacked RELEASE_VERSION fields.
type Payload struct {
	// ReleaseVersion is the releaseVersion protocol field.
	ReleaseVersion string
	// ClientType is the clientType protocol field.
	ClientType string
	// Platform is the platform protocol field.
	Platform int32
	// DeviceCategory is the deviceCategory protocol field.
	DeviceCategory int32
}

// Definition describes the RELEASE_VERSION payload fields.
var Definition = codec.Definition{
	codec.Named("releaseVersion", codec.StringField),
	codec.Named("clientType", codec.StringField),
	codec.Named("platform", codec.Int32Field),
	codec.Named("deviceCategory", codec.Int32Field),
}

// Decode unpacks a RELEASE_VERSION packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}

	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return payloadFromValues(values), nil
}

// payloadFromValues returns a typed payload from decoded values.
func payloadFromValues(values []codec.Value) Payload {
	var payload Payload
	payload.ReleaseVersion = values[0].String
	payload.ClientType = values[1].String
	payload.Platform = values[2].Int32
	payload.DeviceCategory = values[3].Int32

	return payload
}
