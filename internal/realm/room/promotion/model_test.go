package promotion

import (
	"testing"
	"time"
)

// TestPromotionActiveAt verifies the strict expiration boundary.
func TestPromotionActiveAt(t *testing.T) {
	now := time.Unix(100, 0)
	if !(Promotion{ID: 1, EndsAt: now.Add(time.Second)}).ActiveAt(now) {
		t.Fatal("expected active promotion")
	}
	if (Promotion{ID: 1, EndsAt: now}).ActiveAt(now) {
		t.Fatal("expiration boundary must be inactive")
	}
}

// TestValidateCopy verifies Unicode length and required title constraints.
func TestValidateCopy(t *testing.T) {
	if err := validateCopy("Evento", "Descripción"); err != nil {
		t.Fatal(err)
	}
	if err := validateCopy(" ", "x"); err == nil {
		t.Fatal("expected empty-title error")
	}
	if err := validateCopy(string(make([]rune, 65)), "x"); err == nil {
		t.Fatal("expected long-title error")
	}
}

// BenchmarkValidateCopy measures bounded promotion copy validation.
func BenchmarkValidateCopy(b *testing.B) {
	for range b.N {
		_ = validateCopy("Evento QA", "Descripción de sala")
	}
}
