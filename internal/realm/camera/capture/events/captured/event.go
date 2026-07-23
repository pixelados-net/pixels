// Package captured defines the camera capture event.
package captured

const (
	// Name identifies one durable camera capture.
	Name = "camera.captured"
)

// Payload describes one durable camera capture.
type Payload struct {
	// CaptureID identifies the stored capture.
	CaptureID int64
	// PlayerID identifies the capturing player.
	PlayerID int64
	// RoomID identifies the captured room.
	RoomID int64
	// Kind identifies photo or thumbnail capture.
	Kind string
}
