// Package camerafollow contains USER_SETTINGS_CAMERA.
package camerafollow

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_SETTINGS_CAMERA.
const Header uint16 = 1461

// Definition describes the privacy toggle.
var Definition = codec.Definition{codec.Named("cameraFollowBlocked", codec.BooleanField)}

// Payload stores the camera-follow privacy selection.
type Payload struct {
	// CameraFollowBlocked reports whether other players may follow this avatar with the camera.
	CameraFollowBlocked bool
}

// Decode decodes USER_SETTINGS_CAMERA.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{CameraFollowBlocked: values[0].Boolean}, nil
}
