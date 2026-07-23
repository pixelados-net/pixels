package confirm

import (
	"testing"

	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outcompleted "github.com/niflaot/pixels/networking/outbound/trade/completed"
)

// TestCompletionPacketsRefreshInventoryAfterTrade verifies Nitro unlocks stale furniture groups.
func TestCompletionPacketsRefreshInventoryAfterTrade(t *testing.T) {
	packets, err := completionPackets()
	if err != nil {
		t.Fatalf("completion packets: %v", err)
	}
	if packets[0].Header != outcompleted.Header {
		t.Fatalf("first header=%d want=%d", packets[0].Header, outcompleted.Header)
	}
	if packets[1].Header != outrefresh.Header {
		t.Fatalf("second header=%d want=%d", packets[1].Header, outrefresh.Header)
	}
}
