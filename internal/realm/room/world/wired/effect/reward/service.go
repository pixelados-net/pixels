package reward

import (
	"context"
	"strconv"
	"time"

	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playerachievement "github.com/niflaot/pixels/internal/realm/player/achievement"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// Service performs atomic WIRED reward claims and post-commit projection.
type Service struct {
	// store persists reward definitions and claims.
	store record.RewardStore
	// currencies grants configured wallet balances.
	currencies currencyservice.Granter
	// furniture grants inventory items.
	furniture furnitureservice.Granter
	// catalog grants free configured offers.
	catalog catalogservice.Manager
	// achievements grants badges and respect.
	achievements *playerachievement.Service
	// players resolves live connection peers.
	players *playerlive.Registry
	// connections sends reward feedback.
	connections *netconn.Registry
	// now supplies period boundaries.
	now func() time.Time
}

// New creates a reward service.
func New(store record.RewardStore, currencies currencyservice.Granter, furniture furnitureservice.Granter, catalog catalogservice.Manager, achievements *playerachievement.Service, players *playerlive.Registry, connections *netconn.Registry) *Service {
	return &Service{store: store, currencies: currencies, furniture: furniture, catalog: catalog, achievements: achievements, players: players, connections: connections, now: time.Now}
}

// Claim selects and delivers one configured reward.
func (service *Service) Claim(ctx context.Context, node *configuration.Node, event trigger.Event) (effect.Result, error) {
	if node == nil || event.PlayerID <= 0 {
		return effect.Result{Status: effect.Skipped}, nil
	}
	period := periodKey(service.now().UTC(), node.Parameters.Values)
	unique := len(node.Parameters.Values) > 1 && node.Parameters.Values[1] == 1
	status, reward, err := service.store.ClaimReward(ctx, node.ItemID, event.PlayerID, period, unique, event.ID, func(txCtx context.Context, selected record.Reward) error {
		return service.deliver(txCtx, node, event, selected)
	})
	if err != nil {
		return effect.Result{Status: effect.Blocked}, err
	}
	reason := rewardReason(status, reward.Kind, node.Parameters.Values)
	if err = service.send(ctx, event.PlayerID, reason, status == record.ClaimDelivered && reward.Kind != "badge"); err != nil {
		return effect.Result{Status: effect.Applied}, err
	}
	if status != record.ClaimDelivered {
		return effect.Result{Status: effect.Skipped}, nil
	}
	return effect.Result{Status: effect.Applied}, nil
}

// periodKey derives one stable UTC claim window.
func periodKey(now time.Time, values []int32) string {
	kind, interval := int32(0), int32(1)
	if len(values) > 0 {
		kind = values[0]
	}
	if len(values) > 3 && values[3] > 0 {
		interval = values[3]
	}
	switch kind {
	case 1:
		return "day:" + strconv.FormatInt(now.Unix()/(86400*int64(interval)), 10)
	case 2:
		return "hour:" + strconv.FormatInt(now.Unix()/(3600*int64(interval)), 10)
	case 3:
		return "minute:" + strconv.FormatInt(now.Unix()/(60*int64(interval)), 10)
	default:
		return "lifetime"
	}
}

// rewardReason maps durable claim outcomes to Nitro reason codes.
func rewardReason(status record.ClaimStatus, kind string, values []int32) int32 {
	switch status {
	case record.ClaimAlreadyReceived:
		if len(values) > 0 {
			switch values[0] {
			case 1:
				return 2
			case 2:
				return 3
			case 3:
				return 8
			}
		}
		return 1
	case record.ClaimMissed:
		return 4
	case record.ClaimOutOfStock:
		return 5
	case record.ClaimDelivered:
		if kind == "badge" {
			return 7
		}
		return 6
	default:
		return 0
	}
}
