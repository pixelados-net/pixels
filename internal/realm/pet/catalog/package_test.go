package catalog

import (
	"testing"
	"time"
)

// TestPackageRequestValidatesOwnerRoomAndDeadline verifies bounded naming handshakes.
func TestPackageRequestValidatesOwnerRoomAndDeadline(t *testing.T) {
	now := time.Unix(100, 0)
	service := &Service{packageRequests: map[int64]packageRequest{7: {ownerID: 3, roomID: 9, expiresAt: now.Add(time.Minute)}}}
	if !service.validPackageRequest(7, 3, 9, now) {
		t.Fatal("expected current package request")
	}
	if service.validPackageRequest(7, 4, 9, now) || service.validPackageRequest(7, 3, 10, now) || service.validPackageRequest(7, 3, 9, now.Add(time.Minute)) {
		t.Fatal("accepted mismatched or expired package request")
	}
	service.clearPackageRequest(7)
	if service.validPackageRequest(7, 3, 9, now) {
		t.Fatal("completed package request remained active")
	}
}

// TestExpirePackageRequestPreservesReplacement verifies stale callbacks cannot delete new prompts.
func TestExpirePackageRequestPreservesReplacement(t *testing.T) {
	first := time.Unix(100, 0)
	second := first.Add(time.Minute)
	service := &Service{packageRequests: map[int64]packageRequest{7: {ownerID: 3, roomID: 9, expiresAt: second}}}
	service.expirePackageRequest(7, 3, 9, first, second)
	if !service.validPackageRequest(7, 3, 9, first) {
		t.Fatal("stale expiry removed replacement prompt")
	}
	service.expirePackageRequest(7, 3, 9, second, second)
	if service.validPackageRequest(7, 3, 9, first) {
		t.Fatal("matching expiry retained prompt")
	}
}
