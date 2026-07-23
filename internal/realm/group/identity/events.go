package identity

import (
	"context"
	"errors"

	groupobservability "github.com/niflaot/pixels/internal/realm/group/observability"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/pkg/bus"
)

// publish emits one committed identity event when a publisher is configured.
func (service *Service) publish(ctx context.Context, name bus.Name, payload any) {
	if service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: name, Payload: payload})
	}
}

// identityMetricResult classifies low-cardinality identity outcomes.
func identityMetricResult(err error) groupobservability.Result {
	if err == nil {
		return groupobservability.Success
	}
	if errors.Is(err, grouprecord.ErrNotFound) || errors.Is(err, grouprecord.ErrConflict) || errors.Is(err, grouprecord.ErrForbidden) || errors.Is(err, grouprecord.ErrInvalid) || errors.Is(err, grouprecord.ErrLimit) || errors.Is(err, grouprecord.ErrClosed) {
		return groupobservability.Rejected
	}
	return groupobservability.Failed
}
