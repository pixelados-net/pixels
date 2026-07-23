package staff

import (
	"context"
	"time"

	historymodel "github.com/niflaot/pixels/internal/realm/chat/history/model"
	moderationcore "github.com/niflaot/pixels/internal/realm/moderation/core"
	moderationpolicy "github.com/niflaot/pixels/internal/realm/moderation/policy"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incfhlog "github.com/niflaot/pixels/networking/inbound/moderation/staff/cfhchatlog"
	inprefs "github.com/niflaot/pixels/networking/inbound/moderation/staff/preferences"
	inroomlog "github.com/niflaot/pixels/networking/inbound/moderation/staff/roomchatlog"
	inroominfo "github.com/niflaot/pixels/networking/inbound/moderation/staff/roominfo"
	inuserlog "github.com/niflaot/pixels/networking/inbound/moderation/staff/userchatlog"
	"github.com/niflaot/pixels/networking/outbound/moderation/chatrecord"
	outcfhlog "github.com/niflaot/pixels/networking/outbound/moderation/staff/cfhchatlog"
	outprefs "github.com/niflaot/pixels/networking/outbound/moderation/staff/preferences"
	outroomlog "github.com/niflaot/pixels/networking/outbound/moderation/staff/roomchatlog"
	outroominfo "github.com/niflaot/pixels/networking/outbound/moderation/staff/roominfo"
	outuserlog "github.com/niflaot/pixels/networking/outbound/moderation/staff/userchatlog"
)

// preferences sends persisted modtool geometry.
func (handler Handler) preferences(connection netconn.Context, packet codec.Packet) error {
	if err := inprefs.Decode(packet); err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	value, err := handler.Moderation.Store().Preferences(context.Background(), actorID)
	if err != nil {
		return err
	}
	response, err := outprefs.Encode(value.X, value.Y, value.Width, value.Height)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// cfhChatlog sends immutable issue evidence.
func (handler Handler) cfhChatlog(connection netconn.Context, packet codec.Packet) error {
	payload, err := incfhlog.Decode(packet)
	if err != nil {
		return err
	}
	if err = handler.canRead(context.Background(), connection); err != nil {
		return err
	}
	issue, found, err := handler.Moderation.Store().Issue(context.Background(), int64(payload.IssueID), true)
	if err != nil || !found {
		return err
	}
	reported := int32(0)
	if issue.ReportedPlayerID != nil {
		reported = int32(*issue.ReportedPlayerID)
	}
	record := chatRecord(0, "", issue.Chatlog)
	response, err := outcfhlog.Encode(int32(issue.ID), int32(issue.ReporterPlayerID), reported, record)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// roomChatlog sends bounded room history.
func (handler Handler) roomChatlog(connection netconn.Context, packet codec.Packet) error {
	payload, err := inroomlog.Decode(packet)
	if err != nil {
		return err
	}
	if err = handler.canRead(context.Background(), connection); err != nil {
		return err
	}
	roomID := int64(payload.RoomID)
	entries, err := handler.History.History(context.Background(), historymodel.Query{RoomID: &roomID, Limit: 200})
	if err != nil {
		return err
	}
	room, _, _ := handler.Rooms.FindByID(context.Background(), roomID)
	response, err := outroomlog.Encode(historyRecord(roomID, room.Name, entries))
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// userChatlog sends bounded player history grouped as one record.
func (handler Handler) userChatlog(connection netconn.Context, packet codec.Packet) error {
	payload, err := inuserlog.Decode(packet)
	if err != nil {
		return err
	}
	if err = handler.canRead(context.Background(), connection); err != nil {
		return err
	}
	playerID := int64(payload.PlayerID)
	entries, err := handler.History.History(context.Background(), historymodel.Query{PlayerID: &playerID, Limit: 200})
	if err != nil {
		return err
	}
	record, found, err := handler.PlayerRecords.FindByID(context.Background(), playerID)
	if err != nil || !found {
		return err
	}
	response, err := outuserlog.Encode(payload.PlayerID, record.Player.Username, []chatrecord.Record{historyRecord(0, "", entries)})
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// roomInfo sends persistent and live room moderation data.
func (handler Handler) roomInfo(connection netconn.Context, packet codec.Packet) error {
	payload, err := inroominfo.Decode(packet)
	if err != nil {
		return err
	}
	if err = handler.canRead(context.Background(), connection); err != nil {
		return err
	}
	room, found, err := handler.Rooms.FindByID(context.Background(), int64(payload.RoomID))
	if err != nil {
		return err
	}
	params := outroominfo.Params{RoomID: payload.RoomID, Exists: found}
	if found {
		params.OwnerID = int32(room.OwnerPlayerID)
		params.OwnerName = room.OwnerName
		params.Name = room.Name
		params.Description = room.Description
		tags, _ := handler.Rooms.ListTags(context.Background(), room.ID)
		for _, tag := range tags {
			params.Tags = append(params.Tags, tag.Value)
		}
		if active, ok := handler.RoomsLive.Find(room.ID); ok {
			params.UserCount = int32(len(active.Occupants()))
			_, params.OwnerInRoom = active.Unit(room.OwnerPlayerID)
		}
	}
	response, err := outroominfo.Encode(params)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// canRead checks the shared chatlog and visit permission.
func (handler Handler) canRead(ctx context.Context, connection netconn.Context) error {
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	allowed, err := handler.Permissions.HasPermission(ctx, actorID, moderationpolicy.ChatlogRead)
	if err != nil {
		return err
	}
	if !allowed {
		return moderationcore.ErrUnauthorized
	}
	return nil
}

// historyRecord maps durable room chat into Nitro lines.
func historyRecord(roomID int64, roomName string, entries []historymodel.Entry) chatrecord.Record {
	lines := make([]chatrecord.Line, len(entries))
	for index, entry := range entries {
		lines[index] = chatrecord.Line{Timestamp: entry.CreatedAt.Format(time.RFC3339), PlayerID: int32(entry.PlayerID), Message: entry.Message}
	}
	return chatrecord.Record{Type: 0, Context: chatrecord.Context{RoomID: int32(roomID), RoomName: roomName}, Lines: lines}
}

// chatRecord maps frozen issue evidence.
func chatRecord(roomID int64, roomName string, entries []moderationrecord.ChatEntry) chatrecord.Record {
	lines := make([]chatrecord.Line, len(entries))
	for index, entry := range entries {
		playerID := int32(0)
		if entry.PlayerID != nil {
			playerID = int32(*entry.PlayerID)
		}
		lines[index] = chatrecord.Line{Timestamp: entry.CreatedAt.Format(time.RFC3339), PlayerID: playerID, Username: entry.PatternID, Message: entry.Message}
	}
	return chatrecord.Record{Type: 0, Context: chatrecord.Context{RoomID: int32(roomID), RoomName: roomName}, Lines: lines}
}
