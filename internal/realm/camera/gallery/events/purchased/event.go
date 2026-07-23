// Package purchased defines the camera photo purchase event.
package purchased

const (
	// Name identifies one committed photo purchase.
	Name = "camera.purchased"
)

// Payload describes one committed photo purchase.
type Payload struct {
	// CaptureID identifies the consumed capture.
	CaptureID int64
	// PlayerID identifies the buyer.
	PlayerID int64
	// ItemID identifies the granted photo furniture.
	ItemID int64
}
