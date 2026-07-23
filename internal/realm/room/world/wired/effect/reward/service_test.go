package reward

import (
	"context"
	"errors"
	"testing"
	"time"

	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// rewardStore simulates one atomic reward selection.
type rewardStore struct {
	// status stores the selected claim outcome.
	status record.ClaimStatus
	// reward stores the selected reward.
	reward record.Reward
	// err stores an optional persistence failure.
	err error
	// period stores the received period key.
	period string
	// unique stores the received uniqueness policy.
	unique bool
}

// ListRewards is unused by reward service tests.
func (*rewardStore) ListRewards(context.Context, int64) ([]record.Reward, error) { return nil, nil }

// ReplaceRewards is unused by reward service tests.
func (*rewardStore) ReplaceRewards(context.Context, int64, []record.Reward) error { return nil }

// ClaimReward runs delivery for committed outcomes and propagates transaction errors.
func (store *rewardStore) ClaimReward(ctx context.Context, _ int64, _ int64, period string, unique bool, _ uint64, deliver func(context.Context, record.Reward) error) (record.ClaimStatus, record.Reward, error) {
	store.period, store.unique = period, unique
	if store.err != nil {
		return store.status, store.reward, store.err
	}
	if store.status == record.ClaimDelivered {
		if err := deliver(ctx, store.reward); err != nil {
			return store.status, store.reward, err
		}
	}
	return store.status, store.reward, nil
}

// rewardCurrencies records the granted wallet mutation.
type rewardCurrencies struct {
	// params stores the received currency grant.
	params currencyservice.GrantParams
}

// Grant records one wallet mutation.
func (currencies *rewardCurrencies) Grant(_ context.Context, params currencyservice.GrantParams) (int64, error) {
	currencies.params = params
	return params.Amount, nil
}

// TestClaimDeliversCreditsAndUsesStablePeriod verifies atomic selection and post-commit projection.
func TestClaimDeliversCreditsAndUsesStablePeriod(t *testing.T) {
	store := &rewardStore{status: record.ClaimDelivered, reward: record.Reward{Kind: "credits", Reference: "credits", Amount: 10}}
	currencies := &rewardCurrencies{}
	service := New(store, currencies, nil, nil, nil, playerlive.NewRegistry(), nil)
	service.now = func() time.Time { return time.Unix(1721048400, 0) }
	node := &configuration.Node{ItemID: 44, Parameters: configuration.Parameters{Values: []int32{1, 1, 0, 2}}}
	result, err := service.Claim(context.Background(), node, trigger.Event{ID: 9, PlayerID: 7})
	if err != nil || result.Status != effect.Applied {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if currencies.params.PlayerID != 7 || currencies.params.CurrencyType != 0 || currencies.params.Amount != 10 {
		t.Fatalf("currency grant=%+v", currencies.params)
	}
	if store.period == "" || !store.unique {
		t.Fatalf("period=%q unique=%t", store.period, store.unique)
	}
}

// TestClaimMapsSkippedAndBlockedOutcomes verifies invalid actors and persistence errors fail closed.
func TestClaimMapsSkippedAndBlockedOutcomes(t *testing.T) {
	service := New(&rewardStore{}, &rewardCurrencies{}, nil, nil, nil, playerlive.NewRegistry(), nil)
	result, err := service.Claim(context.Background(), nil, trigger.Event{PlayerID: 7})
	if err != nil || result.Status != effect.Skipped {
		t.Fatalf("nil node result=%+v err=%v", result, err)
	}
	result, err = service.Claim(context.Background(), &configuration.Node{}, trigger.Event{})
	if err != nil || result.Status != effect.Skipped {
		t.Fatalf("missing actor result=%+v err=%v", result, err)
	}
	store := &rewardStore{err: errors.New("claim failed")}
	service = New(store, &rewardCurrencies{}, nil, nil, nil, playerlive.NewRegistry(), nil)
	result, err = service.Claim(context.Background(), &configuration.Node{}, trigger.Event{PlayerID: 7})
	if err == nil || result.Status != effect.Blocked {
		t.Fatalf("failed claim result=%+v err=%v", result, err)
	}
	store = &rewardStore{status: record.ClaimAlreadyReceived}
	service = New(store, &rewardCurrencies{}, nil, nil, nil, playerlive.NewRegistry(), nil)
	result, err = service.Claim(context.Background(), &configuration.Node{}, trigger.Event{PlayerID: 7})
	if err != nil || result.Status != effect.Skipped {
		t.Fatalf("duplicate claim result=%+v err=%v", result, err)
	}
}

// TestDeliverRejectsUnsupportedRewardKind verifies custom SQL cannot invoke an undeclared capability.
func TestDeliverRejectsUnsupportedRewardKind(t *testing.T) {
	service := New(&rewardStore{}, &rewardCurrencies{}, nil, nil, nil, playerlive.NewRegistry(), nil)
	err := service.deliver(context.Background(), &configuration.Node{}, trigger.Event{PlayerID: 7}, record.Reward{Kind: "shell", Reference: "1", Amount: 1})
	if err == nil {
		t.Fatal("unsupported reward kind was delivered")
	}
}
