package record

import (
	"context"
	"time"
)

// Store persists call-for-help and modtool state.
type Store interface {
	// Topics lists every topic in display order.
	Topics(context.Context, bool) ([]Topic, error)
	// Topic returns one enabled topic.
	Topic(context.Context, int64) (Topic, bool, error)
	// CreateTopic creates one call-for-help topic.
	CreateTopic(context.Context, Topic) (Topic, error)
	// UpdateTopic replaces one call-for-help topic.
	UpdateTopic(context.Context, Topic) (Topic, bool, error)
	// CreateIssue atomically stores one issue and evidence.
	CreateIssue(context.Context, ReportParams) (Issue, error)
	// Issue returns one issue with optional evidence.
	Issue(context.Context, int64, bool) (Issue, bool, error)
	// Issues lists bounded issues by state.
	Issues(context.Context, string, int32) ([]Issue, error)
	// Pending lists one reporter's open or picked issues.
	Pending(context.Context, int64) ([]Issue, error)
	// DeletePending deletes only one reporter's unresolved issues.
	DeletePending(context.Context, int64) ([]int64, error)
	// Pick atomically claims one open issue.
	Pick(context.Context, int64, int64) (Issue, bool, error)
	// Release returns one moderator-owned issue to the queue.
	Release(context.Context, int64, int64) (Issue, bool, error)
	// Close atomically resolves one assigned issue.
	Close(context.Context, int64, int64, int32) (Issue, bool, error)
	// Preferences returns persisted modtool geometry.
	Preferences(context.Context, int64) (Preferences, error)
	// Presets lists moderator responses.
	Presets(context.Context, bool) ([]Preset, error)
	// CreatePreset creates one moderator response preset.
	CreatePreset(context.Context, Preset) (Preset, error)
	// UpdatePreset replaces one moderator response preset.
	UpdatePreset(context.Context, Preset) (Preset, bool, error)
	// Visits returns a player's recent room visits.
	Visits(context.Context, int64, int32) ([]RoomVisit, error)
	// InsertFeedback stores one completed guide recommendation.
	InsertFeedback(context.Context, int64, int64, bool) error
	// CreateGuardianTicket stores one peer-review ticket.
	CreateGuardianTicket(context.Context, int64, int64, time.Time) (int64, error)
	// SaveGuardianVote records one guardian verdict idempotently.
	SaveGuardianVote(context.Context, int64, int64, int32) error
	// CloseGuardianTicket stores one final peer-review result.
	CloseGuardianTicket(context.Context, int64, int32) error
}
