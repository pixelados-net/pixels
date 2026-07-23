// Package core coordinates call-for-help and moderator issue workflows.
package core

import (
	"context"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
	"github.com/niflaot/pixels/internal/realm/chat/history"
	historymodel "github.com/niflaot/pixels/internal/realm/chat/history/model"
	moderationconfig "github.com/niflaot/pixels/internal/realm/moderation/config"
	issuecreated "github.com/niflaot/pixels/internal/realm/moderation/events/issuecreated"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/redis"
)

// Service coordinates moderation persistence, caching, and authorization.
type Service struct {
	// config stores immutable moderation policy.
	config moderationconfig.Config
	// store persists issues and metadata.
	store moderationrecord.Store
	// history reads bounded hotel chat history.
	history *history.Service
	// throttle stores distributed report counters.
	throttle *redis.Client
	// filter sanitizes visible user text.
	filter *chatfilter.Service
	// permissions resolves staff capabilities.
	permissions permissionservice.Checker
	// events publishes queue changes.
	events bus.Publisher
	// topics stores one immutable enabled snapshot.
	topics atomic.Pointer[[]moderationrecord.Topic]
	// now supplies deterministic timestamps.
	now func() time.Time
}

// New creates moderation behavior.
func New(config moderationconfig.Config, store moderationrecord.Store, history *history.Service, throttle *redis.Client, filter *chatfilter.Service, permissions permissionservice.Checker, events bus.Publisher) *Service {
	return &Service{config: config, store: store, history: history, throttle: throttle, filter: filter, permissions: permissions, events: events, now: time.Now}
}

// Enabled reports call-for-help availability.
func (service *Service) Enabled() bool { return service.config.Enabled }

// RefreshTopics atomically replaces the enabled topic cache.
func (service *Service) RefreshTopics(ctx context.Context) error {
	items, err := service.store.Topics(ctx, false)
	if err != nil {
		return err
	}
	copy := append([]moderationrecord.Topic(nil), items...)
	service.topics.Store(&copy)
	return nil
}

// Topics returns a detached immutable topic snapshot.
func (service *Service) Topics() []moderationrecord.Topic {
	value := service.topics.Load()
	if value == nil {
		return nil
	}
	return append([]moderationrecord.Topic(nil), (*value)...)
}

// Sanitize normalizes and filters user-visible moderation text.
func (service *Service) Sanitize(value string) string { return service.clean(value) }

// ReportResult describes queue, auto-reply, or ignore handling.
type ReportResult struct {
	// Issue stores a queued issue.
	Issue *moderationrecord.Issue
	// ReplyKey stores an automatic localized response.
	ReplyKey string
	// Ignored reports an ignored topic action.
	Ignored bool
}

// Report validates, freezes context, and creates one call for help.
func (service *Service) Report(ctx context.Context, params moderationrecord.ReportParams) (ReportResult, error) {
	if !service.config.Enabled {
		return ReportResult{}, ErrDisabled
	}
	if params.ReporterPlayerID <= 0 || params.TopicID <= 0 {
		return ReportResult{}, ErrInvalid
	}
	if service.throttle != nil {
		count, err := service.throttle.Increment(ctx, "moderation:report:"+strconv.FormatInt(params.ReporterPlayerID, 10), service.config.ReportWindow)
		if err != nil {
			return ReportResult{}, err
		}
		if count > service.config.ReportLimit {
			return ReportResult{}, ErrThrottled
		}
	}
	topic, found := service.topic(params.TopicID)
	if !found {
		return ReportResult{}, ErrNotFound
	}
	params.Message = service.clean(params.Message)
	if topic.Action == "ignore" {
		return ReportResult{Ignored: true}, nil
	}
	if topic.Action == "auto_reply" {
		if topic.AutoReplyKey == nil {
			return ReportResult{Ignored: true}, nil
		}
		return ReportResult{ReplyKey: *topic.AutoReplyKey}, nil
	}
	params.Chatlog = service.freeze(ctx, params)
	issue, err := service.store.CreateIssue(ctx, params)
	if err != nil {
		return ReportResult{}, err
	}
	service.publishCreated(ctx, issue)
	return ReportResult{Issue: &issue}, nil
}

// freeze combines server history with protocol evidence without retaining shared slices.
func (service *Service) freeze(ctx context.Context, params moderationrecord.ReportParams) []moderationrecord.ChatEntry {
	entries := make([]moderationrecord.ChatEntry, 0, service.config.ContextWindow+len(params.Chatlog))
	if service.history != nil {
		query := historymodel.Query{Limit: service.config.ContextWindow, PlayerID: params.ReportedPlayerID, RoomID: params.RoomID}
		if values, err := service.history.History(ctx, query); err == nil {
			for index := len(values) - 1; index >= 0; index-- {
				value := values[index]
				playerID := value.PlayerID
				entries = append(entries, moderationrecord.ChatEntry{PlayerID: &playerID, PatternID: value.Kind, Message: value.Message, CreatedAt: value.CreatedAt})
			}
		}
	}
	for _, entry := range params.Chatlog {
		entry.Message = service.clean(entry.Message)
		entry.PatternID = strings.TrimSpace(entry.PatternID)
		if entry.CreatedAt.IsZero() {
			entry.CreatedAt = service.now()
		}
		entries = append(entries, entry)
	}
	return entries
}

// clean applies protocol-safe trimming and the hotel word filter.
func (service *Service) clean(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if service.filter != nil {
		value, _ = service.filter.Censor(value)
	}
	if len(value) > 500 {
		value = value[:500]
	}
	return value
}

// topic resolves one enabled cached topic.
func (service *Service) topic(id int64) (moderationrecord.Topic, bool) {
	value := service.topics.Load()
	if value == nil {
		return moderationrecord.Topic{}, false
	}
	for _, topic := range *value {
		if topic.ID == id {
			return topic, true
		}
	}
	return moderationrecord.Topic{}, false
}

// publishCreated emits one queue event.
func (service *Service) publishCreated(ctx context.Context, issue moderationrecord.Issue) {
	if service.events == nil {
		return
	}
	reported := int64(0)
	if issue.ReportedPlayerID != nil {
		reported = *issue.ReportedPlayerID
	}
	_ = service.events.Publish(ctx, bus.Event{Name: issuecreated.Name, Payload: issuecreated.Payload{IssueID: issue.ID, ReporterID: issue.ReporterPlayerID, ReportedID: reported, TopicID: issue.TopicID}})
}

// Store returns moderation persistence for focused readers and APIs.
func (service *Service) Store() moderationrecord.Store { return service.store }
