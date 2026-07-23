package tests

import (
	"context"
	"testing"
	"time"

	botbehavior "github.com/niflaot/pixels/internal/realm/bot/behavior"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
)

// BenchmarkBotCycleTick measures one room-owned tick across 25 placed bots.
func BenchmarkBotCycleTick(b *testing.B) {
	registry := botbehavior.NewRegistry()
	if err := botbehavior.RegisterBuiltins(registry); err != nil {
		b.Fatalf("register: %v", err)
	}
	bots := make([]botrecord.Bot, 25)
	for index := range bots {
		bots[index] = placedBot(int64(index+1), botrecord.BehaviorGeneric, false)
		bots[index].X = integerPointer(index % 10)
		bots[index].Y = integerPointer(index / 10)
	}
	service, room := serviceFixture(b, registry, bots)
	if err := service.EnsureRoom(context.Background(), room); err != nil {
		b.Fatalf("ensure room: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if err := service.Cycle(context.Background(), room, time.Now()); err != nil {
			b.Fatalf("cycle: %v", err)
		}
	}
}

// BenchmarkBotUnitMotion isolates allocation-free world reads used by the cycle.
func BenchmarkBotUnitMotion(b *testing.B) {
	registry := botbehavior.NewRegistry()
	_ = botbehavior.RegisterBuiltins(registry)
	service, room := serviceFixture(b, registry, []botrecord.Bot{placedBot(1, botrecord.BehaviorGeneric, false)})
	if err := service.EnsureRoom(context.Background(), room); err != nil {
		b.Fatalf("ensure room: %v", err)
	}
	b.ReportAllocs()
	for range b.N {
		_, _ = room.UnitMotion(-1)
	}
}
