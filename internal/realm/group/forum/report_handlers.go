package forum

import (
	"context"
	"errors"
	"fmt"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inreportmessage "github.com/niflaot/pixels/networking/inbound/group/forum/reportmessage"
	inreportthread "github.com/niflaot/pixels/networking/inbound/group/forum/reportthread"
	outresult "github.com/niflaot/pixels/networking/outbound/moderation/cfh/result"
	"github.com/niflaot/pixels/pkg/i18n"
)

func (handler Handler) reportThread(connection netconn.Context, packet codec.Packet) error {
	if err := inreportthread.Decode(packet); err != nil {
		return err
	}
	cursor, found := handler.Cursors.Get(handler.connectionKey(connection))
	if !found {
		return handler.feedback(connection, "group.forum.report_context_expired")
	}
	return handler.report(connection, cursor, false)
}

// reportMessage validates fresh post report context for moderation intake.
func (handler Handler) reportMessage(connection netconn.Context, packet codec.Packet) error {
	if err := inreportmessage.Decode(packet); err != nil {
		return err
	}
	cursor, found := handler.Cursors.Get(handler.connectionKey(connection))
	if !found || cursor.MessageID <= 0 {
		return handler.feedback(connection, "group.forum.report_context_expired")
	}
	return handler.report(connection, cursor, true)
}

// report freezes authorized forum evidence into the global moderation queue.
func (handler Handler) report(connection netconn.Context, cursor Cursor, message bool) error {
	if handler.Moderation == nil {
		return handler.feedback(connection, "group.forum.report_unavailable")
	}
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	topicID := int64(0)
	for _, topic := range handler.Moderation.Topics() {
		if topic.Category == "forum" {
			topicID = topic.ID
			break
		}
	}
	if topicID == 0 {
		return handler.feedback(connection, "group.forum.report_unavailable")
	}
	evidence := moderationrecord.ChatEntry{PatternID: "GROUP_FORUM_THREAD", CreatedAt: cursor.ViewedAt}
	if message {
		post, postErr := handler.Forum.Post(context.Background(), playerID, cursor.GroupID, cursor.MessageID)
		if postErr != nil {
			return handler.feedback(connection, "group.forum.read_denied")
		}
		evidence.PlayerID = &post.AuthorID
		evidence.PatternID = "GROUP_FORUM_MESSAGE"
		evidence.Message = fmt.Sprintf("group=%d thread=%d message=%d author=%s body=%s", cursor.GroupID, cursor.ThreadID, post.ID, post.AuthorName, post.Body)
	} else {
		thread, _, threadErr := handler.Forum.Thread(context.Background(), playerID, cursor.GroupID, cursor.ThreadID)
		if threadErr != nil {
			return handler.feedback(connection, "group.forum.read_denied")
		}
		evidence.PlayerID = &thread.AuthorID
		evidence.Message = fmt.Sprintf("group=%d thread=%d author=%s subject=%s", cursor.GroupID, thread.ID, thread.AuthorName, thread.Subject)
	}
	result, err := handler.Moderation.Report(context.Background(), moderationrecord.ReportParams{ReporterPlayerID: playerID, ReportedPlayerID: evidence.PlayerID, TopicID: topicID, Kind: "forum", Message: "Forum content report", Chatlog: []moderationrecord.ChatEntry{evidence}})
	if err != nil {
		return handler.feedback(connection, "group.forum.report_failed")
	}
	messageText := handler.translation("moderation.report.received")
	if result.ReplyKey != "" {
		messageText = handler.translation(result.ReplyKey)
	}
	response, _ := outresult.Encode(0, messageText)
	return connection.Send(context.Background(), response)
}

// translation resolves one localized value with a stable key fallback.
func (handler Handler) translation(key string) string {
	if handler.Translations == nil {
		return key
	}
	return handler.Translations.Default(i18n.Key(key))
}

// feedback sends one localized forum error without disconnecting.
func (handler Handler) feedback(connection netconn.Context, key string) error {
	return groupruntime.SendError(context.Background(), connection, handler.Translations, i18n.Key(key))
}

// errorMessages resolves Nitro's five ExtendedForumData error strings.
func (handler Handler) errorMessages() [5]string {
	keys := [5]i18n.Key{"group.forum.read_denied", "group.forum.post_denied", "group.forum.post_denied", "group.forum.post_denied", "group.forum.moderate_denied"}
	var messages [5]string
	for index, key := range keys {
		messages[index] = string(key)
		if handler.Translations != nil {
			messages[index] = handler.Translations.Default(key)
		}
	}
	return messages
}

// statePointer returns an addressable thread state.
func statePointer(state grouprecord.ThreadState) *grouprecord.ThreadState { return &state }

// forumErrorKey maps expected forum failures to localized keys.
func forumErrorKey(err error) string {
	if errors.Is(err, grouprecord.ErrClosed) {
		return "group.forum.thread_locked"
	}
	if errors.Is(err, grouprecord.ErrLimit) {
		return "group.forum.rate_limited"
	}
	return "group.forum.post_denied"
}
