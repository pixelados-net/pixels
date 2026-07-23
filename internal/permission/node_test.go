package permission

import "testing"

// TestNodeValidation verifies concrete and wildcard syntax boundaries.
func TestNodeValidation(t *testing.T) {
	cases := []struct {
		name     string
		node     Node
		valid    bool
		concrete bool
	}{
		{name: "concrete", node: "room.moderation.kick", valid: true, concrete: true},
		{name: "root wildcard", node: "*", valid: true},
		{name: "trailing wildcard", node: "room.moderation.*", valid: true},
		{name: "middle wildcard", node: "room.*.kick"},
		{name: "empty segment", node: "room..kick"},
		{name: "uppercase", node: "Room.kick"},
		{name: "suffix", node: "room."},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.node.Valid() != testCase.valid || testCase.node.Concrete() != testCase.concrete {
				t.Fatalf("unexpected validation for %q", testCase.node)
			}
		})
	}
}

// TestNodeMatchingAndSpecificity verifies wildcard boundaries and precedence.
func TestNodeMatchingAndSpecificity(t *testing.T) {
	query := Node("room.moderation.kick")
	cases := []struct {
		grant       Node
		matches     bool
		specificity int
	}{
		{grant: "room.moderation.kick", matches: true, specificity: 3},
		{grant: "room.moderation.*", matches: true, specificity: 2},
		{grant: "room.*", matches: true, specificity: 1},
		{grant: "*", matches: true, specificity: 0},
		{grant: "room.rights.*", specificity: -1},
		{grant: "room", specificity: -1},
	}
	for _, testCase := range cases {
		if testCase.grant.Matches(query) != testCase.matches || testCase.grant.Specificity(query) != testCase.specificity {
			t.Fatalf("unexpected match for %q", testCase.grant)
		}
	}
}

// BenchmarkNodeValidation measures concrete node syntax validation.
func BenchmarkNodeValidation(b *testing.B) {
	node := Node("room.moderation.any.kick")
	b.ReportAllocs()
	for b.Loop() {
		if !node.Valid() {
			b.Fatal("expected valid node")
		}
	}
}

// BenchmarkNodeSpecificity measures wildcard matching and specificity.
func BenchmarkNodeSpecificity(b *testing.B) {
	grant := Node("room.moderation.*")
	query := Node("room.moderation.any.kick")
	b.ReportAllocs()
	for b.Loop() {
		if grant.Specificity(query) != 2 {
			b.Fatal("unexpected specificity")
		}
	}
}
