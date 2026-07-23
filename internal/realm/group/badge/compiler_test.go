package badge

import (
	"context"
	"testing"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// badgeSource supplies deterministic enabled editor data.
type badgeSource struct{}

// BadgeRegistry returns deterministic enabled editor data.
func (badgeSource) BadgeRegistry(context.Context) ([]grouprecord.BadgeElement, []grouprecord.BadgeColor, error) {
	return []grouprecord.BadgeElement{{Kind: grouprecord.BadgeBase, ID: 1}, {Kind: grouprecord.BadgeSymbol, ID: 12}}, []grouprecord.BadgeColor{{Family: grouprecord.BaseColor, ID: 3, Hex: "FFFFFF"}, {Family: grouprecord.SymbolColor, ID: 162, Hex: "FFFFFF"}}, nil
}

// TestCompileRoundTrip verifies stable renderer-compatible badge code.
func TestCompileRoundTrip(t *testing.T) {
	registry := New(badgeSource{})
	if err := registry.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	compiler := NewCompiler(registry)
	parts := []grouprecord.BadgePart{{Ordinal: 0, Kind: grouprecord.BadgeBase, ElementID: 1, ColorID: 3, Position: 0}, {Ordinal: 1, Kind: grouprecord.BadgeSymbol, ElementID: 12, ColorID: 162, Position: 4}}
	code, normalized, err := compiler.Compile(parts)
	if err != nil || code != "b001030s0121624" || len(normalized) != 2 {
		t.Fatalf("code=%q normalized=%#v err=%v", code, normalized, err)
	}
	parsed, err := Parse(code)
	if err != nil || len(parsed) != 2 || parsed[1] != parts[1] {
		t.Fatalf("parsed=%#v err=%v", parsed, err)
	}
}

// BenchmarkBadgeLookup measures warmed reference lookup.
func BenchmarkBadgeLookup(b *testing.B) {
	registry := New(badgeSource{})
	_ = registry.Refresh(context.Background())
	snapshot, _ := registry.Snapshot()
	b.ReportAllocs()
	for range b.N {
		_ = snapshot.Element(grouprecord.BadgeBase, 1)
	}
}

// BenchmarkBadgeCompile measures bounded mutation-time badge compilation.
func BenchmarkBadgeCompile(b *testing.B) {
	registry := New(badgeSource{})
	_ = registry.Refresh(context.Background())
	compiler := NewCompiler(registry)
	parts := []grouprecord.BadgePart{{Ordinal: 0, Kind: grouprecord.BadgeBase, ElementID: 1, ColorID: 3, Position: 0}, {Ordinal: 1, Kind: grouprecord.BadgeSymbol, ElementID: 12, ColorID: 162, Position: 4}}
	b.ReportAllocs()
	for range b.N {
		_, _, _ = compiler.Compile(parts)
	}
}
