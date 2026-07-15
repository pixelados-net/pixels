package staff

import (
	"context"
	"strconv"
	"time"

	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incloseaction "github.com/niflaot/pixels/networking/inbound/moderation/staff/closeissue"
	inclose "github.com/niflaot/pixels/networking/inbound/moderation/staff/closeissues"
	inpick "github.com/niflaot/pixels/networking/inbound/moderation/staff/pickissues"
	inrelease "github.com/niflaot/pixels/networking/inbound/moderation/staff/releaseissues"
	outresult "github.com/niflaot/pixels/networking/outbound/moderation/cfh/result"
	outdeleted "github.com/niflaot/pixels/networking/outbound/moderation/staff/issuedeleted"
	outinfo "github.com/niflaot/pixels/networking/outbound/moderation/staff/issueinfo"
	outfailed "github.com/niflaot/pixels/networking/outbound/moderation/staff/pickfailed"
)

// pickIssues atomically claims every requested issue independently.
func (handler Handler) pickIssues(connection netconn.Context, packet codec.Packet) error {
	payload, err := inpick.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	for _, issueID := range payload.IssueIDs {
		issue, pickErr := handler.Moderation.Pick(context.Background(), actorID, int64(issueID))
		if pickErr != nil {
			response, _ := outfailed.Encode(issueID, 0, int32(actorID), "", payload.RetryEnabled, payload.RetryCount)
			if sendErr := connection.Send(context.Background(), response); sendErr != nil {
				return sendErr
			}
			continue
		}
		if err = handler.BroadcastIssue(context.Background(), issue.ID); err != nil {
			return err
		}
	}
	return nil
}

// releaseIssues returns moderator-owned issues to the shared queue.
func (handler Handler) releaseIssues(connection netconn.Context, packet codec.Packet) error {
	ids, err := inrelease.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	for _, issueID := range ids {
		issue, releaseErr := handler.Moderation.Release(context.Background(), actorID, int64(issueID))
		if releaseErr != nil {
			continue
		}
		_ = handler.BroadcastIssue(context.Background(), issue.ID)
	}
	return nil
}

// closeIssues resolves moderator-owned issues and notifies reporters.
func (handler Handler) closeIssues(connection netconn.Context, packet codec.Packet) error {
	payload, err := inclose.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	for _, issueID := range payload.IssueIDs {
		issue, closeErr := handler.Moderation.Close(context.Background(), actorID, int64(issueID), payload.Resolution)
		if closeErr != nil {
			return closeErr
		}
		response, _ := outdeleted.Encode(stringID(issue.ID))
		_ = connection.Send(context.Background(), response)
		result, _ := outresult.Encode(payload.Resolution, handler.Text("moderation.report.closed"))
		_ = handler.SendTo(context.Background(), issue.ReporterPlayerID, result)
	}
	return nil
}

// closeDefault resolves and escalates every issue.
func (handler Handler) closeDefault(connection netconn.Context, packet codec.Packet) error {
	payload, err := incloseaction.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	for _, issueID := range payload.IssueIDs {
		issue, closeErr := handler.Moderation.CloseWithDefault(context.Background(), handler.Sanctions, actorID, int64(issueID), 3)
		if closeErr != nil {
			return closeErr
		}
		result, _ := outresult.Encode(3, handler.Text("moderation.report.closed"))
		_ = handler.SendTo(context.Background(), issue.ReporterPlayerID, result)
	}
	return nil
}

// issuePacket maps one durable issue into Nitro wire fields.
func issuePacket(issue moderationrecord.Issue) (codec.Packet, error) {
	return outinfo.Encode(issueParams(issue))
}

// issueParams maps one durable issue into Nitro fields.
func issueParams(issue moderationrecord.Issue) outinfo.Params {
	state := int32(1)
	if issue.State == "picked" {
		state = 2
	} else if issue.State == "resolved" {
		state = 3
	}
	reportedID, pickerID := int32(0), int32(0)
	if issue.ReportedPlayerID != nil {
		reportedID = int32(*issue.ReportedPlayerID)
	}
	if issue.PickedByPlayerID != nil {
		pickerID = int32(*issue.PickedByPlayerID)
	}
	age := int32(time.Since(issue.CreatedAt).Milliseconds())
	return outinfo.Params{IssueID: int32(issue.ID), State: state, CategoryID: int32(issue.TopicID), AgeMilliseconds: age, ReporterID: int32(issue.ReporterPlayerID), ReportedID: reportedID, PickerID: pickerID, Message: issue.Message, ChatRecordID: int32(issue.ID)}
}

// stringID formats a numeric issue id.
func stringID(id int64) string { return strconv.FormatInt(id, 10) }
