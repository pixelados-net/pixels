// Package tests verifies complete canonical effect dispatch.
package tests

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// services records each focused dispatch.
type services struct {
	// calls stores dispatch count.
	calls int
}

// ExecuteFurniture records furniture dispatch.
func (service *services) ExecuteFurniture(context.Context, effect.FurnitureOperation, *configuration.Node, trigger.Event) (effect.Result, error) {
	service.calls++
	return effect.Result{Status: effect.Applied}, nil
}

// ExecuteAvatar records avatar dispatch.
func (service *services) ExecuteAvatar(context.Context, effect.AvatarOperation, *configuration.Node, trigger.Event) (effect.Result, error) {
	service.calls++
	return effect.Result{Status: effect.Applied}, nil
}

// ExecuteBot records bot dispatch.
func (service *services) ExecuteBot(context.Context, effect.BotOperation, *configuration.Node, trigger.Event) (effect.Result, error) {
	service.calls++
	return effect.Result{Status: effect.Applied}, nil
}

// ExecuteGame records game dispatch.
func (service *services) ExecuteGame(context.Context, effect.GameOperation, *configuration.Node, trigger.Event) (effect.Result, error) {
	service.calls++
	return effect.Result{Status: effect.Applied}, nil
}

// Claim records reward dispatch.
func (service *services) Claim(context.Context, *configuration.Node, trigger.Event) (effect.Result, error) {
	service.calls++
	return effect.Result{Status: effect.Applied}, nil
}

// TestAllCanonicalEffectsDispatch verifies no canonical effect is an unhandled no-op.
func TestAllCanonicalEffectsDispatch(t *testing.T) {
	service := &services{}
	executor := effect.New(effect.Services{Furniture: service, Avatar: service, Bot: service, Game: service, Reward: service})
	count := 0
	for _, descriptor := range registry.CanonicalManifest() {
		if descriptor.Family != registry.FamilyEffect {
			continue
		}
		count++
		node := &configuration.Node{Descriptor: descriptor, Targets: []record.Target{{ItemID: 9}}}
		result, err := executor.Execute(context.Background(), node, trigger.Event{})
		if err != nil || result.Status != effect.Applied {
			t.Fatalf("effect %s result = %+v, %v", descriptor.Key, result, err)
		}
	}
	if count != 30 || service.calls != 28 {
		t.Fatalf("effects=%d service calls=%d", count, service.calls)
	}
}
