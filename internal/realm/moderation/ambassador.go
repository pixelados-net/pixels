package moderation

import (
	"context"

	moderationcore "github.com/niflaot/pixels/internal/realm/moderation/core"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
	ambassadoralerted "github.com/niflaot/pixels/internal/realm/room/control/events/ambassadoralerted"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
)

// RegisterAmbassadorIntake routes native ambassador alerts into the moderation issue queue.
func RegisterAmbassadorIntake(lifecycle fx.Lifecycle, subscriber bus.Subscriber, service *moderationcore.Service) error {
	subscription, err := subscriber.Subscribe(ambassadoralerted.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(ambassadoralerted.Payload)
		if !ok {
			return nil
		}
		topicID, found := ambassadorTopic(service.Topics())
		if !found {
			return nil
		}
		reportedID := payload.ReportedPlayerID
		roomID := payload.RoomID
		_, reportErr := service.Report(ctx, moderationrecord.ReportParams{ReporterPlayerID: payload.ReporterPlayerID, ReportedPlayerID: &reportedID, RoomID: &roomID, TopicID: topicID, Kind: "ambassador", Message: "Native ambassador room alert"})
		return reportErr
	})
	if err != nil {
		return err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		subscription.Unsubscribe()
		return nil
	}})
	return nil
}

// ambassadorTopic returns the first enabled queue topic without inventing a parallel intake table.
func ambassadorTopic(topics []moderationrecord.Topic) (int64, bool) {
	for _, topic := range topics {
		if topic.Enabled && topic.Action == "queue" {
			return topic.ID, true
		}
	}
	return 0, false
}
