package membership

import (
	"context"
	"errors"

	changedevent "github.com/niflaot/pixels/internal/realm/group/membership/events/changed"
	favoriteevent "github.com/niflaot/pixels/internal/realm/group/membership/events/favoritechanged"
	groupobservability "github.com/niflaot/pixels/internal/realm/group/observability"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/pkg/bus"
)

// projectChange refreshes projections and publishes one committed membership event.
func (service *Service) projectChange(ctx context.Context, groupID int64, playerID int64, action string) {
	service.refresh(ctx, groupID, playerID)
	if service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: changedevent.Name, Payload: changedevent.Payload{GroupID: groupID, PlayerID: playerID, Action: action}})
	}
}

// publishFavorite emits one committed favorite preference event.
func (service *Service) publishFavorite(ctx context.Context, playerID int64, groupID int64) {
	if service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: favoriteevent.Name, Payload: favoriteevent.Payload{PlayerID: playerID, GroupID: groupID}})
	}
}

// record increments one bounded membership outcome.
func (service *Service) record(kind groupobservability.Kind, err error) {
	service.metrics.Record(groupobservability.Membership, kind, metricResult(err))
}

// metricResult classifies expected social-domain errors without labels.
func metricResult(err error) groupobservability.Result {
	if err == nil {
		return groupobservability.Success
	}
	if errors.Is(err, grouprecord.ErrNotFound) || errors.Is(err, grouprecord.ErrConflict) || errors.Is(err, grouprecord.ErrForbidden) || errors.Is(err, grouprecord.ErrInvalid) || errors.Is(err, grouprecord.ErrLimit) || errors.Is(err, grouprecord.ErrClosed) || errors.Is(err, grouprecord.ErrAlreadyMember) || errors.Is(err, grouprecord.ErrAlreadyPending) {
		return groupobservability.Rejected
	}
	return groupobservability.Failed
}
