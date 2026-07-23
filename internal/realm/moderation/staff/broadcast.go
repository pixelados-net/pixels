package staff

import (
	"context"

	moderationpolicy "github.com/niflaot/pixels/internal/realm/moderation/policy"
	outdeleted "github.com/niflaot/pixels/networking/outbound/moderation/staff/issuedeleted"
)

// RefreshSanctionStatus pushes current sanction state to one online player.
// BroadcastIssue publishes one current issue snapshot only to online moderators.
func (handler Handler) BroadcastIssue(ctx context.Context, issueID int64) error {
	issue, found, err := handler.Moderation.Store().Issue(ctx, issueID, false)
	if err != nil || !found {
		return err
	}
	packet, err := issuePacket(issue)
	if err != nil {
		return err
	}
	return handler.forModerators(ctx, func(playerID int64) error { return handler.SendTo(ctx, playerID, packet) })
}

// BroadcastIssueDeleted removes one resolved issue from every online moderator queue.
func (handler Handler) BroadcastIssueDeleted(ctx context.Context, issueID int64) error {
	packet, err := outdeleted.Encode(stringID(issueID))
	if err != nil {
		return err
	}
	return handler.forModerators(ctx, func(playerID int64) error { return handler.SendTo(ctx, playerID, packet) })
}

// forModerators visits online players currently authorized for the global tool.
func (handler Handler) forModerators(ctx context.Context, visit func(int64) error) error {
	for _, player := range handler.Players.Snapshot() {
		allowed, err := handler.Permissions.HasPermission(ctx, player.ID(), moderationpolicy.ToolAccess)
		if err != nil {
			return err
		}
		if allowed {
			if err = visit(player.ID()); err != nil {
				return err
			}
		}
	}
	return nil
}
