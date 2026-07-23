// Package forum owns social-group forum policies, threads, posts, moderation, and unread state.
package forum

import (
	"context"
	"errors"
	"sync"
	"time"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
	groupconfig "github.com/niflaot/pixels/internal/realm/group/config"
	groupobservability "github.com/niflaot/pixels/internal/realm/group/observability"
	grouppolicy "github.com/niflaot/pixels/internal/realm/group/policy"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	"github.com/niflaot/pixels/pkg/bus"
)

// maxRateEntries triggers stale forum write limiter pruning.
const maxRateEntries = 4096

// Access contains viewer group role and staff override state.
type Access struct {
	// Member reports an active membership.
	Member bool
	// Role stores the active social role when Member is true.
	Role grouprecord.Role
	// Staff reports hotel-wide forum moderation access.
	Staff bool
}

// forumMetricResult classifies low-cardinality forum outcomes.
func forumMetricResult(err error) groupobservability.Result {
	if err == nil {
		return groupobservability.Success
	}
	if errors.Is(err, grouprecord.ErrNotFound) || errors.Is(err, grouprecord.ErrConflict) || errors.Is(err, grouprecord.ErrForbidden) || errors.Is(err, grouprecord.ErrInvalid) || errors.Is(err, grouprecord.ErrLimit) || errors.Is(err, grouprecord.ErrClosed) {
		return groupobservability.Rejected
	}
	return groupobservability.Failed
}

// Service coordinates forum policy and persistence.
type Service struct {
	// config stores page, length, and rate limits.
	config groupconfig.Config
	// store persists forum state.
	store grouprecord.Store
	// players reads author display snapshots.
	players playerservice.Finder
	// permissions resolves hotel staff overrides.
	permissions permissionservice.Checker
	// filter censors configured hotel words.
	filter *chatfilter.Service
	// rateMutex protects the bounded write timestamp map.
	rateMutex sync.Mutex
	// lastPost stores the latest write time per active author.
	lastPost map[int64]time.Time
	// metrics stores bounded process-wide group telemetry.
	metrics *groupobservability.Metrics
	// events publishes committed domain changes.
	events bus.Publisher
}

// New creates group-forum behavior.
func New(config groupconfig.Config, store grouprecord.Store, players playerservice.Finder, permissions permissionservice.Checker, filter *chatfilter.Service, metrics *groupobservability.Metrics, events bus.Publisher) *Service {
	return &Service{config: config, store: store, players: players, permissions: permissions, filter: filter, lastPost: make(map[int64]time.Time), metrics: metrics, events: events}
}

// publish emits one committed forum event when a publisher is configured.
func (service *Service) publish(ctx context.Context, name bus.Name, payload any) {
	if service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: name, Payload: payload})
	}
}

// Stats returns one viewer-authorized forum summary and access state.
func (service *Service) Stats(ctx context.Context, playerID int64, groupID int64) (grouprecord.ForumSummary, Access, error) {
	group, access, err := service.access(ctx, playerID, groupID)
	if err != nil {
		return grouprecord.ForumSummary{}, Access{}, err
	}
	if !allows(group.ReadPolicy, access) {
		return grouprecord.ForumSummary{}, access, grouprecord.ErrForbidden
	}
	summary, found, err := service.store.ForumSummary(ctx, playerID, groupID)
	if err != nil {
		return grouprecord.ForumSummary{}, access, err
	}
	if !found {
		return grouprecord.ForumSummary{}, access, grouprecord.ErrNotFound
	}
	service.metrics.Record(groupobservability.ForumOperations, groupobservability.KindList, groupobservability.Success)
	return summary, access, nil
}

// Summaries returns one bounded authorized forum-list page.
func (service *Service) Summaries(ctx context.Context, playerID int64, mode int32, start int, amount int) ([]grouprecord.ForumSummary, int32, error) {
	started := time.Now()
	defer func() { service.metrics.Observe(groupobservability.ForumQuery, time.Since(started)) }()
	if mode < 0 || mode > 2 || start < 0 || amount < 1 || amount > service.config.ForumPageSize {
		return nil, 0, grouprecord.ErrInvalid
	}
	staff, err := service.has(ctx, playerID, grouppolicy.ForumModerateAny)
	if err != nil {
		return nil, 0, err
	}
	items, total, err := service.store.ForumSummaries(ctx, playerID, mode, start, amount, staff, time.Now().Add(-service.config.ForumActiveWindow))
	service.metrics.Record(groupobservability.ForumOperations, groupobservability.KindList, forumMetricResult(err))
	return items, total, err
}

// Threads returns one authorized bounded thread page.
func (service *Service) Threads(ctx context.Context, playerID int64, groupID int64, start int, amount int) ([]grouprecord.Thread, int32, Access, error) {
	started := time.Now()
	defer func() { service.metrics.Observe(groupobservability.ForumQuery, time.Since(started)) }()
	group, access, err := service.readAccess(ctx, playerID, groupID, start, amount)
	if err != nil {
		return nil, 0, access, err
	}
	threads, total, err := service.store.Threads(ctx, playerID, group.ID, start, amount, access.Staff)
	service.metrics.Record(groupobservability.ForumOperations, groupobservability.KindList, forumMetricResult(err))
	return threads, total, access, err
}

// Posts returns one authorized bounded message page.
func (service *Service) Posts(ctx context.Context, playerID int64, groupID int64, threadID int64, start int, amount int) ([]grouprecord.Post, int32, Access, error) {
	started := time.Now()
	defer func() { service.metrics.Observe(groupobservability.ForumQuery, time.Since(started)) }()
	_, access, err := service.readAccess(ctx, playerID, groupID, start, amount)
	if err != nil {
		return nil, 0, access, err
	}
	posts, total, err := service.store.Posts(ctx, playerID, groupID, threadID, start, amount, access.Staff)
	service.metrics.Record(groupobservability.ForumOperations, groupobservability.KindList, forumMetricResult(err))
	return posts, total, access, err
}

// Thread returns one authorized thread.
func (service *Service) Thread(ctx context.Context, playerID int64, groupID int64, threadID int64) (grouprecord.Thread, Access, error) {
	group, access, err := service.access(ctx, playerID, groupID)
	if err != nil || !allows(group.ReadPolicy, access) {
		return grouprecord.Thread{}, access, grouprecord.ErrForbidden
	}
	thread, found, err := service.store.Thread(ctx, groupID, threadID, access.Staff)
	if err == nil && !found {
		err = grouprecord.ErrNotFound
	}
	return thread, access, err
}

// Post returns one authorized retained post for moderation or report context.
func (service *Service) Post(ctx context.Context, playerID int64, groupID int64, postID int64) (grouprecord.Post, error) {
	group, access, err := service.access(ctx, playerID, groupID)
	if err != nil || !allows(group.ReadPolicy, access) {
		return grouprecord.Post{}, grouprecord.ErrForbidden
	}
	post, found, err := service.store.Post(ctx, groupID, postID)
	if err == nil && !found {
		err = grouprecord.ErrNotFound
	}
	return post, err
}

// CreateThread creates a thread and first post atomically.
