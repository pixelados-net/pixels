package main

import (
	"testing"
)

// TestNewAppBuilds verifies the dependency graph can be constructed.
func TestNewAppBuilds(t *testing.T) {
	app := newApp()

	if app == nil {
		t.Fatal("expected app")
	}
}

// TestOptionsBuilds verifies dependency graph options are registered.
func TestOptionsBuilds(t *testing.T) {
	options := options()

	if len(options) != 10 {
		t.Fatalf("expected ten options, got %d", len(options))
	}
}
