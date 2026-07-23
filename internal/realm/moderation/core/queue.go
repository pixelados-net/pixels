package core

import (
	"context"

	"github.com/niflaot/pixels/internal/permission"
	issueclosed "github.com/niflaot/pixels/internal/realm/moderation/events/issueclosed"
	moderationpolicy "github.com/niflaot/pixels/internal/realm/moderation/policy"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
	sanctioncore "github.com/niflaot/pixels/internal/realm/sanction/core"
	"github.com/niflaot/pixels/pkg/bus"
)

// Pick atomically claims one open issue.
func (service *Service) Pick(ctx context.Context, moderatorID int64, issueID int64) (moderationrecord.Issue, error) {
	if err := service.authorize(ctx, moderatorID, moderationpolicy.IssueManage); err != nil {
		return moderationrecord.Issue{}, err
	}
	issue, updated, err := service.store.Pick(ctx, issueID, moderatorID)
	if err != nil {
		return moderationrecord.Issue{}, err
	}
	if !updated {
		return moderationrecord.Issue{}, ErrPickFailed
	}
	return issue, nil
}

// Release returns one assigned issue to the queue.
func (service *Service) Release(ctx context.Context, moderatorID int64, issueID int64) (moderationrecord.Issue, error) {
	if err := service.authorize(ctx, moderatorID, moderationpolicy.IssueManage); err != nil {
		return moderationrecord.Issue{}, err
	}
	issue, updated, err := service.store.Release(ctx, issueID, moderatorID)
	if err != nil {
		return moderationrecord.Issue{}, err
	}
	if !updated {
		return moderationrecord.Issue{}, ErrNotFound
	}
	return issue, nil
}

// Close resolves one issue and publishes its result.
func (service *Service) Close(ctx context.Context, moderatorID int64, issueID int64, resolution int32) (moderationrecord.Issue, error) {
	if err := service.authorize(ctx, moderatorID, moderationpolicy.IssueManage); err != nil {
		return moderationrecord.Issue{}, err
	}
	issue, updated, err := service.store.Close(ctx, issueID, moderatorID, resolution)
	if err != nil {
		return moderationrecord.Issue{}, err
	}
	if !updated {
		return moderationrecord.Issue{}, ErrNotFound
	}
	if service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: issueclosed.Name, Payload: issueclosed.Payload{IssueID: issue.ID, Resolution: resolution, ModeratorID: moderatorID}})
	}
	return issue, nil
}

// CloseWithDefault resolves and escalates the reported player.
func (service *Service) CloseWithDefault(ctx context.Context, sanctions *sanctioncore.Service, moderatorID int64, issueID int64, resolution int32) (moderationrecord.Issue, error) {
	issue, err := service.Close(ctx, moderatorID, issueID, resolution)
	if err != nil || issue.ReportedPlayerID == nil {
		return issue, err
	}
	_, err = sanctions.EscalateFor(ctx, sanctioncore.EscalateParams{ReceiverPlayerID: *issue.ReportedPlayerID, Reason: issue.Message, IssueID: &issue.ID})
	return issue, err
}

// authorize checks one staff capability.
func (service *Service) authorize(ctx context.Context, playerID int64, node permission.Node) error {
	allowed, err := service.permissions.HasPermission(ctx, playerID, node)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrUnauthorized
	}
	return nil
}
