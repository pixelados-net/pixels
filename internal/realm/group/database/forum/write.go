package forum

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// CreateThread atomically inserts a thread and first post.
func (repository *Repository) CreateThread(ctx context.Context, groupID int64, authorID int64, authorName string, authorFigure string, subject string, body string) (grouprecord.Thread, grouprecord.Post, error) {
	var thread grouprecord.Thread
	var post grouprecord.Post
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := repository.executor(txCtx).QueryRow(txCtx, `insert into social_group_forum_threads(group_id,author_player_id,author_name,subject,post_count,last_author_player_id,last_author_name) values($1,$2,$3,$4,1,$2,$3) returning id,group_id,author_player_id,author_name,subject,state,pinned,locked,post_count,last_post_id,last_author_player_id,last_author_name,last_posted_at,moderator_player_id,moderator_name,moderation_reason,moderated_at,created_at,updated_at,version`, groupID, authorID, authorName, subject).Scan(&thread.ID, &thread.GroupID, &thread.AuthorID, &thread.AuthorName, &thread.Subject, &thread.State, &thread.Pinned, &thread.Locked, &thread.PostCount, &thread.LastPostID, &thread.LastAuthorID, &thread.LastAuthorName, &thread.LastPostedAt, &thread.ModeratorID, &thread.ModeratorName, &thread.ModerationReason, &thread.ModeratedAt, &thread.CreatedAt, &thread.UpdatedAt, &thread.Version); err != nil {
			return err
		}
		created, err := repository.insertPost(txCtx, groupID, thread.ID, 0, authorID, authorName, authorFigure, body)
		if err != nil {
			return err
		}
		post = created
		if _, err = repository.executor(txCtx).Exec(txCtx, `update social_group_forum_threads set last_post_id=$2,last_posted_at=$3 where id=$1`, thread.ID, post.ID, post.CreatedAt); err != nil {
			return err
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `update social_groups set thread_count=thread_count+1,post_count=post_count+1,updated_at=now(),version=version+1 where id=$1 and forum_enabled and deactivated_at is null`, groupID); err != nil {
			return err
		}
		thread.LastPostID, thread.LastPostedAt = post.ID, post.CreatedAt
		return nil
	})
	return thread, post, err
}

// CreatePost atomically inserts one reply and advances counters.
func (repository *Repository) CreatePost(ctx context.Context, groupID int64, threadID int64, authorID int64, authorName string, authorFigure string, body string) (grouprecord.Post, error) {
	var post grouprecord.Post
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		var ordinal int32
		var locked bool
		var state grouprecord.ThreadState
		if err := repository.executor(txCtx).QueryRow(txCtx, `select post_count,locked,state from social_group_forum_threads where id=$1 and group_id=$2 for update`, threadID, groupID).Scan(&ordinal, &locked, &state); err != nil {
			return grouprecord.ErrNotFound
		}
		if locked || state != grouprecord.ThreadOpen {
			return grouprecord.ErrClosed
		}
		created, err := repository.insertPost(txCtx, groupID, threadID, ordinal, authorID, authorName, authorFigure, body)
		if err != nil {
			return err
		}
		post = created
		if _, err = repository.executor(txCtx).Exec(txCtx, `update social_group_forum_threads set post_count=post_count+1,last_post_id=$2,last_author_player_id=$3,last_author_name=$4,last_posted_at=$5,updated_at=now(),version=version+1 where id=$1`, threadID, post.ID, authorID, authorName, post.CreatedAt); err != nil {
			return err
		}
		_, err = repository.executor(txCtx).Exec(txCtx, `update social_groups set post_count=post_count+1,updated_at=now(),version=version+1 where id=$1`, groupID)
		return err
	})
	return post, err
}

// insertPost inserts one post with a transaction-assigned ordinal.
func (repository *Repository) insertPost(ctx context.Context, groupID int64, threadID int64, ordinal int32, authorID int64, authorName string, authorFigure string, body string) (grouprecord.Post, error) {
	return scanPost(repository.executor(ctx).QueryRow(ctx, `insert into social_group_forum_posts(group_id,thread_id,ordinal,author_player_id,author_name,author_figure,body) values($1,$2,$3,$4,$5,$6,$7) returning id,group_id,thread_id,ordinal,author_player_id,author_name,author_figure,body,state,moderator_player_id,moderator_name,moderation_reason,moderated_at,(select count(*) from social_group_forum_posts authored where authored.group_id=$1 and authored.author_player_id=$4)+1,created_at,updated_at,version`, groupID, threadID, ordinal, authorID, authorName, authorFigure, body))
}

// UpdateThread changes pin, lock, or moderation state optimistically.
func (repository *Repository) UpdateThread(ctx context.Context, groupID int64, threadID int64, version int64, pinned *bool, locked *bool, state *grouprecord.ThreadState, moderatorID int64, reason string) (grouprecord.Thread, error) {
	var thread grouprecord.Thread
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		command, err := repository.executor(txCtx).Exec(txCtx, `update social_group_forum_threads set pinned=coalesce($4,pinned),locked=coalesce($5,locked),state=coalesce($6,state),moderator_player_id=case when $6 is null then moderator_player_id else $7 end,moderator_name=case when $6 is null then moderator_name else coalesce((select username from players where id=$7),'') end,moderation_reason=case when $6 is null then moderation_reason else $8 end,moderated_at=case when $6 is null then moderated_at else now() end,updated_at=now(),version=version+1 where id=$1 and group_id=$2 and version=$3`, threadID, groupID, version, pinned, locked, state, moderatorID, reason)
		if err != nil {
			return err
		}
		if command.RowsAffected() != 1 {
			return grouprecord.ErrConflict
		}
		var found bool
		thread, found, err = repository.Thread(txCtx, groupID, threadID, true)
		if err == nil && !found {
			return grouprecord.ErrNotFound
		}
		if err != nil {
			return err
		}
		return repository.audit(txCtx, groupID, "group.forum.thread.updated", thread.Version)
	})
	return thread, err
}

// UpdatePost changes retained post moderation state optimistically.
func (repository *Repository) UpdatePost(ctx context.Context, groupID int64, postID int64, version int64, state grouprecord.PostState, moderatorID int64, reason string) (grouprecord.Post, error) {
	var post grouprecord.Post
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		command, err := repository.executor(txCtx).Exec(txCtx, `update social_group_forum_posts set state=$4,moderator_player_id=$5,moderator_name=coalesce((select username from players where id=$5),''),moderation_reason=$6,moderated_at=now(),updated_at=now(),version=version+1 where id=$1 and group_id=$2 and version=$3`, postID, groupID, version, state, moderatorID, reason)
		if err != nil {
			return err
		}
		if command.RowsAffected() != 1 {
			return grouprecord.ErrConflict
		}
		post, err = scanPost(repository.executor(txCtx).QueryRow(txCtx, `select post.id,post.group_id,post.thread_id,post.ordinal,post.author_player_id,post.author_name,post.author_figure,post.body,post.state,post.moderator_player_id,post.moderator_name,post.moderation_reason,post.moderated_at,(select count(*) from social_group_forum_posts authored where authored.group_id=post.group_id and authored.author_player_id=post.author_player_id),post.created_at,post.updated_at,post.version from social_group_forum_posts post where post.id=$1`, postID))
		if err != nil {
			return err
		}
		return repository.audit(txCtx, groupID, "group.forum.post.updated", post.Version)
	})
	return post, err
}

// audit writes attributed forum administration inside the current transaction.
func (repository *Repository) audit(ctx context.Context, groupID int64, action string, version int64) error {
	attribution, found := grouprecord.AuditFromContext(ctx)
	if !found {
		return nil
	}
	_, err := repository.executor(ctx).Exec(ctx, `insert into social_group_audit(group_id,actor_player_id,action,reason,version) values($1,$2,$3,$4,$5)`, groupID, attribution.ActorPlayerID, action, attribution.Reason, version)
	return err
}

// UpdateReadMarker advances one marker monotonically.
func (repository *Repository) UpdateReadMarker(ctx context.Context, marker grouprecord.ReadMarker) (grouprecord.ReadMarker, error) {
	err := repository.executor(ctx).QueryRow(ctx, `insert into social_group_forum_read_markers(group_id,player_id,last_message_id,flag) values($1,$2,$3,$4) on conflict(group_id,player_id) do update set last_message_id=greatest(social_group_forum_read_markers.last_message_id,excluded.last_message_id),flag=excluded.flag,updated_at=now() returning group_id,player_id,last_message_id,flag,updated_at`, marker.GroupID, marker.PlayerID, marker.LastMessageID, marker.Flag).Scan(&marker.GroupID, &marker.PlayerID, &marker.LastMessageID, &marker.Flag, &marker.UpdatedAt)
	return marker, err
}

// Post returns one retained post for report context.
func (repository *Repository) Post(ctx context.Context, groupID int64, postID int64) (grouprecord.Post, bool, error) {
	post, err := scanPost(repository.executor(ctx).QueryRow(ctx, `select post.id,post.group_id,post.thread_id,post.ordinal,post.author_player_id,post.author_name,post.author_figure,post.body,post.state,post.moderator_player_id,post.moderator_name,post.moderation_reason,post.moderated_at,(select count(*) from social_group_forum_posts authored where authored.group_id=post.group_id and authored.author_player_id=post.author_player_id),post.created_at,post.updated_at,post.version from social_group_forum_posts post where post.group_id=$1 and post.id=$2`, groupID, postID))
	if errors.Is(err, pgx.ErrNoRows) {
		return grouprecord.Post{}, false, nil
	}
	return post, err == nil, err
}
