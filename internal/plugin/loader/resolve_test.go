package loader

import (
	"errors"
	"reflect"
	"testing"

	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
)

// resolveCase stores one dependency graph expectation.
type resolveCase struct {
	// name identifies the graph case.
	name string
	// candidates stores the fixture graph.
	candidates []candidate
	// blocked stores the expected blocked plugin count.
	blocked int
}

// TestResolveOrdersDependencies verifies dependency-first topological ordering.
func TestResolveOrdersDependencies(t *testing.T) {
	candidates := []candidate{
		testCandidate("child", "base"),
		testCandidate("base"),
	}
	ordered, blocked := resolve(candidates)
	names := candidateNames(ordered)
	if len(blocked) != 0 || !reflect.DeepEqual(names, []string{"base", "child"}) {
		t.Fatalf("expected dependency order, names=%v blocked=%v", names, blocked)
	}
}

// TestResolveRejectsDuplicatesMissingDependenciesAndCycles verifies isolated graph failures.
func TestResolveRejectsDuplicatesMissingDependenciesAndCycles(t *testing.T) {
	tests := []resolveCase{
		{name: "duplicate", candidates: []candidate{testCandidate("same"), testCandidate("same")}, blocked: 1},
		{name: "missing", candidates: []candidate{testCandidate("child", "missing"), testCandidate("healthy")}, blocked: 1},
		{name: "cycle", candidates: []candidate{testCandidate("one", "two"), testCandidate("two", "one"), testCandidate("healthy")}, blocked: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ordered, blocked := resolve(test.candidates)
			if len(blocked) != test.blocked {
				t.Fatalf("expected %d blocked, got %v", test.blocked, blocked)
			}
			if test.name != "duplicate" && !reflect.DeepEqual(candidateNames(ordered), []string{"healthy"}) {
				t.Fatalf("expected healthy plugin isolation, got %v", candidateNames(ordered))
			}
			for _, err := range blocked {
				if !errors.Is(err, ErrDependency) && !errors.Is(err, ErrInvalidMetadata) {
					t.Fatalf("expected classified graph error, got %v", err)
				}
			}
		})
	}
}

// testCandidate creates one valid graph fixture.
func testCandidate(name string, dependencies ...string) candidate {
	return candidate{metadata: sdkplugin.Metadata{Name: name, Version: "1.0.0", Author: "QA", SDKVersion: sdkplugin.SDKVersion, Dependencies: dependencies}}
}

// candidateNames returns ordered manifest names.
func candidateNames(candidates []candidate) []string {
	names := make([]string, 0, len(candidates))
	for _, current := range candidates {
		names = append(names, current.metadata.Name)
	}
	return names
}
