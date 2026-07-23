package forum

import (
	"context"
	"time"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inmoderatemessage "github.com/niflaot/pixels/networking/inbound/group/forum/moderatemessage"
	inmoderatethread "github.com/niflaot/pixels/networking/inbound/group/forum/moderatethread"
	inpost "github.com/niflaot/pixels/networking/inbound/group/forum/post"
	inmarker "github.com/niflaot/pixels/networking/inbound/group/forum/readmarker"
	inreportmessage "github.com/niflaot/pixels/networking/inbound/group/forum/reportmessage"
	inreportthread "github.com/niflaot/pixels/networking/inbound/group/forum/reportthread"
	insettings "github.com/niflaot/pixels/networking/inbound/group/forum/settings"
	inupdatethread "github.com/niflaot/pixels/networking/inbound/group/forum/updatethread"
	outpost "github.com/niflaot/pixels/networking/outbound/group/forum/post"
	outpostthread "github.com/niflaot/pixels/networking/outbound/group/forum/postthread"
	outmessage "github.com/niflaot/pixels/networking/outbound/group/forum/updatemessage"
	outthread "github.com/niflaot/pixels/networking/outbound/group/forum/updatethread"
)

// registerWriteHandlers registers forum mutations and report cursors.
func registerWriteHandlers(registry *netconn.HandlerRegistry, handler Handler) {
	_ = registry.Register(inpost.Header, handler.post)
	_ = registry.Register(inmoderatemessage.Header, handler.moderateMessage)
	_ = registry.Register(inmoderatethread.Header, handler.moderateThread)
	_ = registry.Register(inmarker.Header, handler.readMarkers)
	_ = registry.Register(insettings.Header, handler.settings)
	_ = registry.Register(inupdatethread.Header, handler.updateThread)
	_ = registry.Register(inreportthread.Header, handler.reportThread)
	_ = registry.Register(inreportmessage.Header, handler.reportMessage)
}

// post creates a new thread or replies to an existing thread.
func (handler Handler) post(connection netconn.Context, packet codec.Packet) error {
	payload, err := inpost.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	if payload.ThreadID == 0 {
		thread, _, createErr := handler.Forum.CreateThread(context.Background(), playerID, payload.GroupID, payload.Subject, payload.Message)
		if createErr != nil {
			return handler.feedback(connection, forumErrorKey(createErr))
		}
		response, encodeErr := outpostthread.Encode(payload.GroupID, thread, time.Now())
		if encodeErr != nil {
			return encodeErr
		}
		return handler.project(connection, playerID, payload.GroupID, response)
	}
	created, err := handler.Forum.CreatePost(context.Background(), playerID, payload.GroupID, payload.ThreadID, payload.Message)
	if err != nil {
		return handler.feedback(connection, forumErrorKey(err))
	}
	response, err := outpost.Encode(payload.GroupID, payload.ThreadID, created, time.Now())
	if err != nil {
		return err
	}
	return handler.project(connection, playerID, payload.GroupID, response)
}

// moderateMessage changes one retained message visibility state.
func (handler Handler) moderateMessage(connection netconn.Context, packet codec.Packet) error {
	payload, err := inmoderatemessage.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	current, err := handler.Forum.Post(context.Background(), playerID, payload.GroupID, payload.MessageID)
	if err != nil {
		return handler.feedback(connection, "group.forum.read_denied")
	}
	updated, err := handler.Forum.UpdatePost(context.Background(), playerID, payload.GroupID, payload.MessageID, current.Version, grouprecord.PostState(payload.State), "nitro moderation")
	if err != nil {
		return handler.feedback(connection, forumErrorKey(err))
	}
	response, err := outmessage.Encode(payload.GroupID, payload.ThreadID, updated, time.Now())
	if err != nil {
		return err
	}
	return handler.project(connection, playerID, payload.GroupID, response)
}

// moderateThread changes one retained thread visibility state.
func (handler Handler) moderateThread(connection netconn.Context, packet codec.Packet) error {
	payload, err := inmoderatethread.Decode(packet)
	if err != nil {
		return err
	}
	return handler.mutateThread(connection, payload.GroupID, payload.ThreadID, nil, nil, statePointer(grouprecord.ThreadState(payload.State)))
}

// updateThread changes the pinned and locked flags in composer order.
func (handler Handler) updateThread(connection netconn.Context, packet codec.Packet) error {
	payload, err := inupdatethread.Decode(packet)
	if err != nil {
		return err
	}
	return handler.mutateThread(connection, payload.GroupID, payload.ThreadID, &payload.Pinned, &payload.Locked, nil)
}

// mutateThread resolves the current version and sends the updated projection.
func (handler Handler) mutateThread(connection netconn.Context, groupID int64, threadID int64, pinned *bool, locked *bool, state *grouprecord.ThreadState) error {
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	current, _, err := handler.Forum.Thread(context.Background(), playerID, groupID, threadID)
	if err != nil {
		return handler.feedback(connection, "group.forum.read_denied")
	}
	updated, err := handler.Forum.UpdateThread(context.Background(), playerID, groupID, threadID, current.Version, pinned, locked, state, "nitro moderation")
	if err != nil {
		return handler.feedback(connection, forumErrorKey(err))
	}
	response, err := outthread.Encode(groupID, updated, time.Now())
	if err != nil {
		return err
	}
	return handler.project(connection, playerID, groupID, response)
}

// project sends one mutation to its actor and every other fresh forum viewer.
func (handler Handler) project(connection netconn.Context, actorID int64, groupID int64, packet codec.Packet) error {
	if err := connection.Send(context.Background(), packet); err != nil {
		return err
	}
	for _, playerID := range handler.Cursors.Viewers(groupID) {
		if playerID != actorID {
			_, _ = handler.Delivery.Send(context.Background(), playerID, packet)
		}
	}
	return nil
}

// readMarkers advances one bounded marker collection monotonically.
func (handler Handler) readMarkers(connection netconn.Context, packet codec.Packet) error {
	markers, err := inmarker.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	_, err = handler.Forum.UpdateReadMarkers(context.Background(), playerID, markers)
	if err != nil {
		return handler.feedback(connection, "group.forum.read_denied")
	}
	return nil
}

// settings changes the four forum access policies.
func (handler Handler) settings(connection netconn.Context, packet codec.Packet) error {
	payload, err := insettings.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	summary, _, err := handler.Forum.Stats(context.Background(), playerID, payload.GroupID)
	if err != nil {
		return handler.feedback(connection, "group.forum.read_denied")
	}
	_, err = handler.Forum.UpdateSettings(context.Background(), playerID, payload.GroupID, summary.Group.Version, true, grouprecord.Policy(payload.ReadPolicy), grouprecord.Policy(payload.PostMessagePolicy), grouprecord.Policy(payload.PostThreadPolicy), grouprecord.Policy(payload.ModeratePolicy))
	if err != nil {
		return handler.feedback(connection, forumErrorKey(err))
	}
	return nil
}

// reportThread validates fresh report context for moderation intake.
