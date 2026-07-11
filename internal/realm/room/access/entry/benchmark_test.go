package entry

import (
	"context"
	"testing"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/pkg/i18n"
	"github.com/niflaot/pixels/pkg/model"
)

// BenchmarkAuthorizeOpen measures the dominant room entry policy path.
func BenchmarkAuthorizeOpen(b *testing.B) {
	service := New(Config{}, nil, nil, nil, Nodes{})
	request := Request{Room: roommodel.Room{Base: model.Base{Identity: model.Identity{ID: 9}}, OwnerPlayerID: 7, DoorMode: roommodel.DoorModeOpen}, PlayerID: 8}
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_, _ = service.Authorize(ctx, request)
	}
}

// BenchmarkEntryKey measures Redis key construction allocations.
func BenchmarkEntryKey(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_ = attemptKey(922337203685477580, 922337203685477581)
	}
}

// BenchmarkLockoutAlert measures the precomputed user-facing lockout message.
func BenchmarkLockoutAlert(b *testing.B) {
	catalog := i18n.NewCatalog(i18n.Config{DefaultLocale: "en"}, map[i18n.Locale]map[i18n.Key]string{
		"en": {"room.entry.locked": "Locked for {duration}.", "duration.minute.other": "{count} minutes"},
	})
	service := New(Config{LockoutSeconds: 600}, nil, nil, catalog, Nodes{})
	b.ReportAllocs()
	for b.Loop() {
		_ = service.lockoutAlert()
	}
}
