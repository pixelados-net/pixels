// Package data appends shared Nitro group-forum records.
package data

import (
	"time"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
)

// AppendSummary appends one ForumData record.
func AppendSummary(dst []byte, summary grouprecord.ForumSummary, now time.Time) ([]byte, error) {
	lastID := int64(0)
	lastAuthorID := int64(0)
	lastAuthorName := ""
	lastSeconds := int32(0)
	if summary.LastPost != nil {
		lastID = summary.LastPost.ID
		lastAuthorID = summary.LastPost.AuthorID
		lastAuthorName = summary.LastPost.AuthorName
		lastSeconds = secondsAgo(now, summary.LastPost.CreatedAt)
	}
	return codec.AppendPayload(dst, codec.Definition{
		codec.Int32Field, codec.StringField, codec.StringField, codec.StringField,
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field,
		codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field,
	}, codec.Int32(int32(summary.Group.ID)), codec.String(summary.Group.Name), codec.String(summary.Group.Description), codec.String(summary.Group.BadgeCode),
		codec.Int32(summary.Group.ThreadCount), codec.Int32(summary.LeaderboardScore), codec.Int32(summary.Group.PostCount), codec.Int32(summary.UnreadMessages),
		codec.Int32(int32(lastID)), codec.Int32(int32(lastAuthorID)), codec.String(lastAuthorName), codec.Int32(lastSeconds))
}

// AppendThread appends one GuildForumThread record.
func AppendThread(dst []byte, thread grouprecord.Thread, now time.Time) ([]byte, error) {
	moderatorID := int64(0)
	if thread.ModeratorID != nil {
		moderatorID = *thread.ModeratorID
	}
	moderatedAgo := int32(0)
	if thread.ModeratedAt != nil {
		moderatedAgo = secondsAgo(now, *thread.ModeratedAt)
	}
	return codec.AppendPayload(dst, codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.StringField, codec.StringField, codec.BooleanField, codec.BooleanField,
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField,
		codec.Int32Field, codec.ByteField, codec.Int32Field, codec.StringField, codec.Int32Field,
	}, codec.Int32(int32(thread.ID)), codec.Int32(int32(thread.AuthorID)), codec.String(thread.AuthorName), codec.String(thread.Subject), codec.Bool(thread.Pinned), codec.Bool(thread.Locked),
		codec.Int32(secondsAgo(now, thread.CreatedAt)), codec.Int32(thread.PostCount), codec.Int32(thread.UnreadCount), codec.Int32(int32(thread.LastPostID)), codec.Int32(int32(thread.LastAuthorID)), codec.String(thread.LastAuthorName),
		codec.Int32(secondsAgo(now, thread.LastPostedAt)), codec.Byte(byte(thread.State)), codec.Int32(int32(moderatorID)), codec.String(thread.ModeratorName), codec.Int32(moderatedAgo))
}

// AppendPost appends one MessageData record.
func AppendPost(dst []byte, post grouprecord.Post, now time.Time) ([]byte, error) {
	moderatorID := int64(0)
	if post.ModeratorID != nil {
		moderatorID = *post.ModeratorID
	}
	moderatedAgo := int32(0)
	if post.ModeratedAt != nil {
		moderatedAgo = secondsAgo(now, *post.ModeratedAt)
	}
	return codec.AppendPayload(dst, codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.StringField, codec.Int32Field,
		codec.StringField, codec.ByteField, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field,
	}, codec.Int32(int32(post.ID)), codec.Int32(post.Ordinal), codec.Int32(int32(post.AuthorID)), codec.String(post.AuthorName), codec.String(post.AuthorFigure), codec.Int32(secondsAgo(now, post.CreatedAt)),
		codec.String(post.Body), codec.Byte(byte(post.State)), codec.Int32(int32(moderatorID)), codec.String(post.ModeratorName), codec.Int32(moderatedAgo), codec.Int32(post.AuthorPostCount))
}

// secondsAgo converts one timestamp to a non-negative bounded duration.
func secondsAgo(now time.Time, value time.Time) int32 {
	if value.IsZero() || value.After(now) {
		return 0
	}
	seconds := now.Sub(value) / time.Second
	if seconds > time.Duration(^uint32(0)>>1) {
		return int32(^uint32(0) >> 1)
	}
	return int32(seconds)
}
