// Package published defines the camera publication event.
package published

const (
	// Name identifies one committed camera publication.
	Name = "camera.published"
)

// Payload describes one committed camera publication.
type Payload struct {
	// PublicationID identifies the gallery entry.
	PublicationID int64
	// CaptureID identifies the consumed capture.
	CaptureID int64
	// PlayerID identifies the publisher.
	PlayerID int64
}
