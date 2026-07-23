package handlers

import (
	"context"

	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
	moderationruntime "github.com/niflaot/pixels/internal/realm/moderation/runtime"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	ininvite "github.com/niflaot/pixels/networking/inbound/moderation/guide/invite"
	inreport "github.com/niflaot/pixels/networking/inbound/moderation/guide/report"
	instatus "github.com/niflaot/pixels/networking/inbound/moderation/guide/reportingstatus"
	inrequesterroom "github.com/niflaot/pixels/networking/inbound/moderation/guide/requesterroom"
	outdetached "github.com/niflaot/pixels/networking/outbound/moderation/guide/detached"
	outinvited "github.com/niflaot/pixels/networking/outbound/moderation/guide/invited"
	outstatus "github.com/niflaot/pixels/networking/outbound/moderation/guide/reportingstatus"
	outrequesterroom "github.com/niflaot/pixels/networking/outbound/moderation/guide/requesterroom"
)

// Disconnected removes a player from guide state and informs the partner.
func Disconnected(ctx context.Context, runtime *moderationruntime.Context, playerID int64) error {
	session, found := runtime.Guides.RemovePlayer(playerID)
	if !found {
		return nil
	}
	partner, found := session.Partner(playerID)
	if !found {
		return nil
	}
	packet, _ := outdetached.Encode()
	return runtime.SendTo(ctx, partner, packet)
}

// guideInvite sends the guide room metadata to the requester.
func (handler Handler) guideInvite(connection netconn.Context, packet codec.Packet) error {
	if err := ininvite.Decode(packet); err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	session, found := handler.Guides.SessionFor(actorID)
	if !found || session.GuidePlayerID != actorID {
		return nil
	}
	player, found := handler.Players.Find(actorID)
	if !found {
		return nil
	}
	roomID, found := player.CurrentRoom()
	if !found {
		return nil
	}
	room, _, _ := handler.Rooms.FindByID(context.Background(), roomID)
	response, _ := outinvited.Encode(int32(roomID), room.Name)
	return handler.SendTo(context.Background(), session.RequesterPlayerID, response)
}

// guideRequesterRoom sends the requester's current room id to the guide.
func (handler Handler) guideRequesterRoom(connection netconn.Context, packet codec.Packet) error {
	if err := inrequesterroom.Decode(packet); err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	session, found := handler.Guides.SessionFor(actorID)
	if !found || session.GuidePlayerID != actorID {
		return nil
	}
	requester, found := handler.Players.Find(session.RequesterPlayerID)
	if !found {
		return nil
	}
	roomID, _ := requester.CurrentRoom()
	response, _ := outrequesterroom.Encode(int32(roomID))
	return connection.Send(context.Background(), response)
}

// guideReport escalates the active transcript into the moderator queue.
func (handler Handler) guideReport(connection netconn.Context, packet codec.Packet) error {
	payload, err := inreport.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	session, found := handler.Guides.SessionFor(actorID)
	if !found {
		return nil
	}
	target, topicID := session.RequesterPlayerID, int64(1)
	params := moderationrecord.ReportParams{ReporterPlayerID: actorID, ReportedPlayerID: &target, TopicID: topicID, Kind: "guide", Message: payload.Reason}
	for _, message := range session.Transcript {
		playerID := message.SenderPlayerID
		params.Chatlog = append(params.Chatlog, moderationrecord.ChatEntry{PlayerID: &playerID, Message: message.Text, CreatedAt: message.CreatedAt})
	}
	_, err = handler.Moderation.Report(context.Background(), params)
	return err
}

// guideStatus reports whether the actor currently has a reportable session.
func (handler Handler) guideStatus(connection netconn.Context, packet codec.Packet) error {
	if err := instatus.Decode(packet); err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	_, found := handler.Guides.SessionFor(actorID)
	response, _ := outstatus.Encode(0, 0, 1, found, "", "", "", "")
	return connection.Send(context.Background(), response)
}
