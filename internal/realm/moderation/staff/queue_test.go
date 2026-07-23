package staff

import (
	"testing"
	"time"

	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
	outissueclose "github.com/niflaot/pixels/networking/outbound/moderation/staff/issueclose"
)

// TestIssueParamsProjectsPlayerNames verifies visible ticket identities are not blanked.
func TestIssueParamsProjectsPlayerNames(t *testing.T) {
	reported, picker := int64(3), int64(1)
	params := issueParams(moderationrecord.Issue{ID: 9, ReporterPlayerID: 4, ReporterName: "Carol", ReportedPlayerID: &reported, ReportedName: "Bob", TopicID: 1, State: "picked", PickedByPlayerID: &picker, PickerName: "Demo", Message: "Help", CreatedAt: time.Now()})
	if params.ReporterName != "Carol" || params.ReportedName != "Bob" || params.PickerName != "Demo" || params.ReporterID != 4 || params.ReportedID != 3 || params.PickerID != 1 {
		t.Fatalf("params=%+v", params)
	}
}

// TestIssueClosePacketUsesNativeClosureHeader guards against the former CFH-result transport bug.
func TestIssueClosePacketUsesNativeClosureHeader(t *testing.T) {
	packet, err := issueClosePacket(2, "cerrado")
	if err != nil || packet.Header != outissueclose.Header {
		t.Fatalf("header=%d err=%v", packet.Header, err)
	}
}
