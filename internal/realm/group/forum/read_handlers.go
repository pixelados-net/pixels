package forum

import (
	"context"
	"fmt"
	"time"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
	moderationcore "github.com/niflaot/pixels/internal/realm/moderation/core"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inlist "github.com/niflaot/pixels/networking/inbound/group/forum/list"
	inmessages "github.com/niflaot/pixels/networking/inbound/group/forum/messages"
	instats "github.com/niflaot/pixels/networking/inbound/group/forum/stats"
	inthread "github.com/niflaot/pixels/networking/inbound/group/forum/thread"
	inthreads "github.com/niflaot/pixels/networking/inbound/group/forum/threads"
	inunread "github.com/niflaot/pixels/networking/inbound/group/forum/unread"
	outlist "github.com/niflaot/pixels/networking/outbound/group/forum/list"
	outmessages "github.com/niflaot/pixels/networking/outbound/group/forum/messages"
	outstats "github.com/niflaot/pixels/networking/outbound/group/forum/stats"
	outthreads "github.com/niflaot/pixels/networking/outbound/group/forum/threads"
	outunread "github.com/niflaot/pixels/networking/outbound/group/forum/unreadcount"
	"github.com/niflaot/pixels/pkg/i18n"
)

// Handler contains forum packet adapter dependencies.
type Handler struct {
	// Forum manages forum policy and durable content.
	Forum *Service
	// Cursors stores header-only report context.
	Cursors *Cursors
	// Delivery resolves authenticated actors.
	Delivery *groupruntime.Delivery
	// Translations localizes policy errors.
	Translations i18n.Translator
	// Moderation accepts frozen forum report evidence.
	Moderation *moderationcore.Service
}

// RegisterHandlers registers all consumed group-forum packets.
func RegisterHandlers(registry *netconn.HandlerRegistry, handler Handler) {
	registerReadHandlers(registry, handler)
	registerWriteHandlers(registry, handler)
}

// registerReadHandlers registers forum query adapters.
func registerReadHandlers(registry *netconn.HandlerRegistry, handler Handler) {
	_ = registry.Register(instats.Header, handler.stats)
	_ = registry.Register(inlist.Header, handler.list)
	_ = registry.Register(inthreads.Header, handler.threads)
	_ = registry.Register(inmessages.Header, handler.messages)
	_ = registry.Register(inthread.Header, handler.thread)
	_ = registry.Register(inunread.Header, handler.unread)
}

// stats sends one ExtendedForumData response.
func (handler Handler) stats(connection netconn.Context, packet codec.Packet) error {
	groupID, err := instats.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	summary, access, err := handler.Forum.Stats(context.Background(), playerID, groupID)
	if err != nil {
		return handler.feedback(connection, "group.forum.read_denied")
	}
	handler.Cursors.Set(handler.connectionKey(connection), Cursor{PlayerID: playerID, GroupID: groupID})
	errors := handler.errorMessages()
	canChange := access.Staff || access.Member && access.Role == 0
	response, err := outstats.Encode(summary, errors, canChange, access.Staff, time.Now())
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// list sends one bounded forum summary page.
func (handler Handler) list(connection netconn.Context, packet codec.Packet) error {
	payload, err := inlist.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	items, total, err := handler.Forum.Summaries(context.Background(), playerID, payload.Mode, int(payload.Start), int(payload.Amount))
	if err != nil {
		return handler.feedback(connection, "group.forum.read_denied")
	}
	response, err := outlist.Encode(payload.Mode, total, payload.Start, items, time.Now())
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// threads sends one bounded forum thread page.
func (handler Handler) threads(connection netconn.Context, packet codec.Packet) error {
	payload, err := inthreads.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	items, _, _, err := handler.Forum.Threads(context.Background(), playerID, payload.GroupID, int(payload.Start), int(payload.Amount))
	if err != nil {
		return handler.feedback(connection, "group.forum.read_denied")
	}
	handler.Cursors.Set(handler.connectionKey(connection), Cursor{PlayerID: playerID, GroupID: payload.GroupID})
	response, err := outthreads.Encode(payload.GroupID, payload.Start, items, time.Now())
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// messages sends one bounded forum post page and records report context.
func (handler Handler) messages(connection netconn.Context, packet codec.Packet) error {
	payload, err := inmessages.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	items, _, _, err := handler.Forum.Posts(context.Background(), playerID, payload.GroupID, payload.ThreadID, int(payload.Start), int(payload.Amount))
	if err != nil {
		return handler.feedback(connection, "group.forum.read_denied")
	}
	messageID := int64(0)
	if len(items) > 0 {
		messageID = items[len(items)-1].ID
	}
	handler.Cursors.Set(handler.connectionKey(connection), Cursor{PlayerID: playerID, GroupID: payload.GroupID, ThreadID: payload.ThreadID, MessageID: messageID})
	response, err := outmessages.Encode(payload.GroupID, payload.ThreadID, payload.Start, items, time.Now())
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// thread refreshes one selected thread through the standard thread list shape.
func (handler Handler) thread(connection netconn.Context, packet codec.Packet) error {
	payload, err := inthread.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	item, _, err := handler.Forum.Thread(context.Background(), playerID, payload.GroupID, payload.ThreadID)
	if err != nil {
		return handler.feedback(connection, "group.forum.read_denied")
	}
	handler.Cursors.Set(handler.connectionKey(connection), Cursor{PlayerID: playerID, GroupID: payload.GroupID, ThreadID: payload.ThreadID})
	response, err := outthreads.Encode(payload.GroupID, 0, []grouprecord.Thread{item}, time.Now())
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// unread sends one hotel-wide authorized unread count.
func (handler Handler) unread(connection netconn.Context, packet codec.Packet) error {
	if err := inunread.Decode(packet); err != nil {
		return err
	}
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	count, err := handler.Forum.UnreadCount(context.Background(), playerID)
	if err != nil {
		return err
	}
	response, err := outunread.Encode(count)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// connectionKey identifies one transport session for report cursors.
func (handler Handler) connectionKey(connection netconn.Context) string {
	return fmt.Sprintf("%s:%s", connection.ConnectionKind, connection.ConnectionID)
}
