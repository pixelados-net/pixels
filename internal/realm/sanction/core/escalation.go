package core

import (
	"context"
	"time"

	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
)

// EscalateParams describes one automatic ladder sanction.
type EscalateParams struct {
	// ReceiverPlayerID identifies the target.
	ReceiverPlayerID int64
	// Reason stores the originating issue explanation.
	Reason string
	// CFHTopicID optionally links a topic.
	CFHTopicID *int64
	// IssueID optionally links an issue.
	IssueID *int64
	// Source overrides the default escalation source.
	Source string
}

// EscalateFor resolves the next probation level through the common engine.
func (service *Service) EscalateFor(ctx context.Context, params EscalateParams) (sanctionrecord.Punishment, error) {
	ladder, err := service.store.Ladder(ctx)
	if err != nil {
		return sanctionrecord.Punishment{}, err
	}
	if len(ladder) == 0 {
		return sanctionrecord.Punishment{}, ErrLadderEmpty
	}
	level := int32(1)
	last, lastLevel, found, err := service.store.LastEscalation(ctx, params.ReceiverPlayerID)
	if err != nil {
		return sanctionrecord.Punishment{}, err
	}
	if found {
		entry := ladderEntry(ladder, lastLevel)
		if service.now().Before(last.IssuedAt.Add(time.Duration(entry.ProbationDays) * 24 * time.Hour)) {
			level = lastLevel + 1
		}
	}
	entry := ladderEntry(ladder, level)
	var expiresAt *time.Time
	if entry.DurationHours > 0 {
		value := service.now().Add(time.Duration(entry.DurationHours) * time.Hour)
		expiresAt = &value
	}
	source := params.Source
	if source == "" {
		source = "escalation"
	}
	return service.Apply(ctx, sanctionrecord.ApplyParams{ReceiverPlayerID: params.ReceiverPlayerID, IssuerKind: "system", Kind: entry.Kind, Reason: params.Reason, CFHTopicID: params.CFHTopicID, IssueID: params.IssueID, Source: source, ExpiresAt: expiresAt})
}

// ladderEntry returns the requested or highest configured level.
func ladderEntry(entries []sanctionrecord.LadderEntry, level int32) sanctionrecord.LadderEntry {
	selected := entries[len(entries)-1]
	for _, entry := range entries {
		if entry.Level >= level {
			return entry
		}
	}
	return selected
}
