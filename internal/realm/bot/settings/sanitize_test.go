package settings

import "testing"

// TestSanitizeFixedPointStripsNestedEncodedHTML verifies defensive convergence.
func TestSanitizeFixedPointStripsNestedEncodedHTML(t *testing.T) {
	value, converged := sanitizeFixedPoint("&amp;lt;b&amp;gt;hello&amp;lt;/b&amp;gt;")
	if !converged || value != "hello" {
		t.Fatalf("value=%q converged=%t", value, converged)
	}
	value, converged = sanitizeFixedPoint("&amp;amp;amp;amp;amp;amp;lt;b&amp;amp;amp;amp;amp;amp;gt;x")
	if converged || value != "" {
		t.Fatalf("expected defensive empty result, value=%q converged=%t", value, converged)
	}
}

// TestTruncateRunesPreservesUnicode verifies code-point limits.
func TestTruncateRunesPreservesUnicode(t *testing.T) {
	if value := truncateRunes("áéí", 2); value != "áé" {
		t.Fatalf("unexpected value %q", value)
	}
}
