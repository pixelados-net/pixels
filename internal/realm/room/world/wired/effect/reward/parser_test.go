package reward

import (
	"reflect"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
)

// TestParseSupportsEveryDurableRewardKind verifies compact editor normalization.
func TestParseSupportsEveryDurableRewardKind(t *testing.T) {
	rows, err := Parse("0,ach_wiredqa1,2;1,furni#17,3;1,cata#22,4;1,credits#10,5;1,points7#9,6;1,respect#2,7")
	if err != nil {
		t.Fatal(err)
	}
	wantKinds := []string{"badge", "furniture", "catalog_offer", "credits", "currency", "respect"}
	for index, want := range wantKinds {
		if rows[index].Kind != want || rows[index].Ordinal != index {
			t.Fatalf("row %d=%+v, want kind %s", index, rows[index], want)
		}
	}
	if rows[0].Reference != "ACH_WIREDQA1" || rows[3].Amount != 10 || rows[4].Reference != "7" || rows[4].Amount != 9 {
		t.Fatalf("unexpected normalized rows: %+v", rows)
	}
}

// TestParseRejectsInvalidEntries verifies malformed rewards fail before persistence.
func TestParseRejectsInvalidEntries(t *testing.T) {
	for _, value := range []string{"", "1", "1,furni#2,0", "0,,1", "1,unknown#2,1", "1,points#zero,1"} {
		if _, err := Parse(value); err == nil {
			t.Fatalf("Parse(%q) succeeded", value)
		}
	}
}

// TestPeriodKeysAndReasonsCoverNitroCodes verifies periods and all client reason codes.
func TestPeriodKeysAndReasonsCoverNitroCodes(t *testing.T) {
	now := time.Unix(1721048400, 0).UTC()
	keys := []string{
		periodKey(now, []int32{0}),
		periodKey(now, []int32{1, 0, 0, 2}),
		periodKey(now, []int32{2}),
		periodKey(now, []int32{3}),
	}
	if keys[0] != "lifetime" || keys[1] == keys[2] || keys[2] == keys[3] {
		t.Fatalf("period keys=%v", keys)
	}
	codes := []int32{
		rewardReason(record.ClaimUnavailable, "", nil),
		rewardReason(record.ClaimAlreadyReceived, "", nil),
		rewardReason(record.ClaimAlreadyReceived, "", []int32{1}),
		rewardReason(record.ClaimAlreadyReceived, "", []int32{2}),
		rewardReason(record.ClaimAlreadyReceived, "", []int32{3}),
		rewardReason(record.ClaimMissed, "", nil),
		rewardReason(record.ClaimOutOfStock, "", nil),
		rewardReason(record.ClaimDelivered, "furniture", nil),
		rewardReason(record.ClaimDelivered, "badge", nil),
	}
	if !reflect.DeepEqual(codes, []int32{0, 1, 2, 3, 8, 4, 5, 6, 7}) {
		t.Fatalf("reason codes=%v", codes)
	}
}
