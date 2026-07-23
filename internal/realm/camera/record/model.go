// Package record defines camera persistence records and domain contracts.
package record

import (
	"encoding/json"
	"strings"
	"time"
)

// Kind identifies one stored camera artifact.
type Kind string

// State identifies the durable lifecycle of one photo capture.
type State string

const (
	// KindPhoto identifies a purchasable room photo.
	KindPhoto Kind = "photo"
	// KindThumbnail identifies a room navigator thumbnail.
	KindThumbnail Kind = "thumbnail"
	// StatePending identifies the current editable photo.
	StatePending State = "pending"
	// StatePurchased identifies an active photo with purchased furniture.
	StatePurchased State = "purchased"
	// StatePublished identifies an active photo published to the gallery.
	StatePublished State = "published"
	// StatePurchasedPublished identifies an active purchased and published photo.
	StatePurchasedPublished State = "purchased_published"
	// StateSuperseded identifies a photo replaced by a newer capture.
	StateSuperseded State = "superseded"
	// StateAbandoned identifies an unreferenced photo claimed for cleanup.
	StateAbandoned State = "abandoned"
	// StateDeleted identifies an object removed from durable storage.
	StateDeleted State = "deleted"
)

// Capture stores one uploaded camera artifact.
type Capture struct {
	// ID identifies the capture row.
	ID int64 `json:"id"`
	// UUID stores the durable client photo identifier.
	UUID string `json:"uuid"`
	// PlayerID identifies the capturing player.
	PlayerID int64 `json:"playerId"`
	// RoomID identifies the captured room.
	RoomID int64 `json:"roomId"`
	// Kind identifies the artifact purpose.
	Kind Kind `json:"kind"`
	// State identifies the photo lifecycle.
	State State `json:"state"`
	// StorageKey stores the object key.
	StorageKey string `json:"storageKey"`
	// URL stores the permanent public object URL.
	URL string `json:"url"`
	// CreatedAt stores the upload time.
	CreatedAt time.Time `json:"createdAt"`
	// ConsumedAt stores the legacy replacement or cleanup time.
	ConsumedAt *time.Time `json:"consumedAt,omitempty"`
	// SupersededAt stores when a newer photo replaced this capture.
	SupersededAt *time.Time `json:"supersededAt,omitempty"`
	// AbandonedAt stores when cleanup claimed this capture.
	AbandonedAt *time.Time `json:"abandonedAt,omitempty"`
	// DeletedAt stores when object storage deletion completed.
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
	// CleanupAttemptedAt stores the latest cleanup claim time.
	CleanupAttemptedAt *time.Time `json:"cleanupAttemptedAt,omitempty"`
	// PurchaseCount stores the number of furniture copies created.
	PurchaseCount int `json:"purchaseCount"`
	// Version stores optimistic state.
	Version int64 `json:"version"`
}

// CleanupCandidate identifies one unreferenced object claimed for deletion.
type CleanupCandidate struct {
	// CaptureID identifies the claimed capture.
	CaptureID int64
	// StorageKey identifies the object to delete.
	StorageKey string
}

// PhotoCompanionKey returns Nitro's wall-rendering object key for one photo.
func PhotoCompanionKey(storageKey string) (string, bool) {
	if !strings.HasSuffix(storageKey, ".png") {
		return "", false
	}
	return strings.TrimSuffix(storageKey, ".png") + "_small.png", true
}

// Publication stores one public gallery entry.
type Publication struct {
	// ID identifies the publication.
	ID int64 `json:"id"`
	// CaptureID identifies the source capture.
	CaptureID int64 `json:"captureId"`
	// PlayerID identifies the publisher.
	PlayerID int64 `json:"playerId"`
	// RoomID identifies the captured room.
	RoomID int64 `json:"roomId"`
	// URL stores the permanent image URL.
	URL string `json:"url"`
	// CreatedAt stores the publication time.
	CreatedAt time.Time `json:"createdAt"`
	// RemovedAt stores an optional moderation time.
	RemovedAt *time.Time `json:"removedAt,omitempty"`
	// RemovedReason stores an optional moderation reason.
	RemovedReason string `json:"removedReason,omitempty"`
}

// Settings stores the singleton operational camera policy.
type Settings struct {
	// Enabled controls camera operations.
	Enabled bool
	// CreditsPrice stores the purchase credit price.
	CreditsPrice int64
	// PointsPrice stores the purchase points price.
	PointsPrice int64
	// PointsType identifies the purchase points currency.
	PointsType int32
	// PublishPointsPrice stores the publication price.
	PublishPointsPrice int64
	// PublishPointsType identifies the publication currency.
	PublishPointsType int32
	// PublishCooldown stores the publication cooldown.
	PublishCooldown time.Duration
	// UpdatedAt stores the last settings mutation.
	UpdatedAt time.Time
	// Version stores optimistic settings state.
	Version int64
}

// PhotoData stores Nitro's external-image furniture payload.
type PhotoData struct {
	// Timestamp stores the capture Unix timestamp.
	Timestamp int64 `json:"t"`
	// UUID stores the unique capture identifier.
	UUID string `json:"u"`
	// RoomID stores the source room identifier.
	RoomID int64 `json:"s"`
	// URL stores the permanent image URL.
	URL string `json:"w"`
	// Caption stores the optional photo caption.
	Caption string `json:"m"`
	// Name stores the photo owner display name.
	Name string `json:"n"`
	// OwnerName stores the photo owner profile name.
	OwnerName string `json:"o"`
	// OwnerID stores the photo owner identifier.
	OwnerID int64 `json:"oi"`
}

// JSON encodes Nitro's compact external-image payload.
func (data PhotoData) JSON() (string, error) {
	encoded, err := json.Marshal(data)
	return string(encoded), err
}

// Audit stores durable administrative attribution.
type Audit struct {
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64
	// Action stores the bounded mutation name.
	Action string
	// EntityID identifies the affected row.
	EntityID int64
	// Reason stores the required human-readable attribution.
	Reason string
}
