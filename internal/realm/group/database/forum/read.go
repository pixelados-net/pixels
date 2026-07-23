package forum

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// ForumSummaries lists visible forum summaries by protocol mode.
func (repository *Repository) ForumSummaries(ctx context.Context, playerID int64, mode int32, offset int, limit int, staff bool, activeSince time.Time) ([]grouprecord.ForumSummary, int32, error) {
	filter := `g.deactivated_at is null and g.forum_enabled`
	order := `g.updated_at desc,g.id desc`
	if mode == 1 {
		order = `coalesce(popularity.views,0) desc,g.updated_at desc,g.id desc`
	}
	if mode == 0 {
		filter += ` and g.updated_at >= $5`
	}
	if mode == 2 {
		filter += ` and membership.player_id is not null`
	}
	filter += ` and ($4 or g.forum_read_policy=0 or (membership.player_id is not null and (g.forum_read_policy=1 or (g.forum_read_policy=2 and membership.role<=1) or (g.forum_read_policy=3 and membership.role=0))))`
	filter += ` and $2>=0 and $3>0`
	base := ` from social_groups g left join social_group_members membership on membership.group_id=g.id and membership.player_id=$1 left join player_social_group_preferences preference on preference.player_id=$1 left join social_group_forum_read_markers marker on marker.group_id=g.id and marker.player_id=$1 left join lateral (select count(*)::int views from social_group_forum_views view where view.group_id=g.id and view.viewed_on>=$5::date) popularity on true left join lateral (select p.id,p.author_player_id,p.author_name,p.author_figure,p.body,p.state,p.created_at from social_group_forum_posts p where p.group_id=g.id and p.state=0 order by p.id desc limit 1) last on true where ` + filter
	var total int32
	if err := repository.executor(ctx).QueryRow(ctx, `select count(*)`+base, playerID, offset, limit, staff, activeSince).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := repository.executor(ctx).Query(ctx, `select g.id,g.owner_player_id,g.name,g.description,g.home_room_id,g.state,g.can_members_decorate,g.color_a,g.color_b,g.badge_code,g.member_count,g.pending_count,g.thread_count,g.post_count,g.created_at,g.updated_at,g.version,g.forum_read_policy,g.forum_post_message_policy,g.forum_post_thread_policy,g.forum_moderate_policy,coalesce(popularity.views,0),coalesce((select count(*) from social_group_forum_posts unread where unread.group_id=g.id and unread.state=0 and unread.id>coalesce(marker.last_message_id,0)),0),last.id,last.author_player_id,last.author_name,last.author_figure,last.body,last.state,last.created_at`+base+` order by `+order+` offset $2 limit $3`, playerID, offset, limit, staff, activeSince)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	summaries := make([]grouprecord.ForumSummary, 0, limit)
	for rows.Next() {
		var summary grouprecord.ForumSummary
		var lastID, lastAuthorID *int64
		var lastName, lastFigure, lastBody *string
		var lastState *grouprecord.PostState
		var lastCreated *time.Time
		err = rows.Scan(&summary.Group.ID, &summary.Group.OwnerPlayerID, &summary.Group.Name, &summary.Group.Description, &summary.Group.HomeRoomID,
			&summary.Group.State, &summary.Group.CanMembersDecorate, &summary.Group.ColorA, &summary.Group.ColorB, &summary.Group.BadgeCode,
			&summary.Group.MemberCount, &summary.Group.PendingCount, &summary.Group.ThreadCount, &summary.Group.PostCount,
			&summary.Group.CreatedAt, &summary.Group.UpdatedAt, &summary.Group.Version, &summary.Group.ReadPolicy, &summary.Group.PostMessagePolicy, &summary.Group.PostThreadPolicy, &summary.Group.ModeratePolicy, &summary.LeaderboardScore, &summary.UnreadMessages,
			&lastID, &lastAuthorID, &lastName, &lastFigure, &lastBody, &lastState, &lastCreated)
		if err != nil {
			return nil, 0, err
		}
		summary.Group.ForumEnabled = true
		if lastID != nil {
			summary.LastPost = &grouprecord.Post{ID: *lastID, GroupID: summary.Group.ID, AuthorID: *lastAuthorID, AuthorName: *lastName, AuthorFigure: *lastFigure, Body: *lastBody, State: *lastState, CreatedAt: *lastCreated}
		}
		summaries = append(summaries, summary)
	}
	return summaries, total, rows.Err()
}

// ForumSummary returns one viewer-specific forum summary.
func (repository *Repository) ForumSummary(ctx context.Context, playerID int64, groupID int64) (grouprecord.ForumSummary, bool, error) {
	rows, _, err := repository.forumSummariesForGroup(ctx, playerID, groupID)
	if err != nil || len(rows) == 0 {
		return grouprecord.ForumSummary{}, false, err
	}
	return rows[0], true, nil
}

// forumSummariesForGroup returns one summary using the list projection contract.
func (repository *Repository) forumSummariesForGroup(ctx context.Context, playerID int64, groupID int64) ([]grouprecord.ForumSummary, int32, error) {
	rows, err := repository.executor(ctx).Query(ctx, `select g.id,g.owner_player_id,g.name,g.description,g.home_room_id,g.state,g.can_members_decorate,g.color_a,g.color_b,g.badge_code,g.member_count,g.pending_count,g.thread_count,g.post_count,g.created_at,g.updated_at,g.version,g.forum_read_policy,g.forum_post_message_policy,g.forum_post_thread_policy,g.forum_moderate_policy,0,coalesce((select count(*) from social_group_forum_posts unread where unread.group_id=g.id and unread.state=0 and unread.id>coalesce(marker.last_message_id,0)),0),last.id,last.author_player_id,last.author_name,last.author_figure,last.body,last.state,last.created_at from social_groups g left join social_group_forum_read_markers marker on marker.group_id=g.id and marker.player_id=$1 left join lateral (select p.id,p.author_player_id,p.author_name,p.author_figure,p.body,p.state,p.created_at from social_group_forum_posts p where p.group_id=g.id and p.state=0 order by p.id desc limit 1) last on true where g.id=$2 and g.deactivated_at is null and g.forum_enabled`, playerID, groupID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, 0, rows.Err()
	}
	var summary grouprecord.ForumSummary
	var lastID, lastAuthorID *int64
	var lastName, lastFigure, lastBody *string
	var lastState *grouprecord.PostState
	var lastCreated *time.Time
	err = rows.Scan(&summary.Group.ID, &summary.Group.OwnerPlayerID, &summary.Group.Name, &summary.Group.Description, &summary.Group.HomeRoomID,
		&summary.Group.State, &summary.Group.CanMembersDecorate, &summary.Group.ColorA, &summary.Group.ColorB, &summary.Group.BadgeCode,
		&summary.Group.MemberCount, &summary.Group.PendingCount, &summary.Group.ThreadCount, &summary.Group.PostCount,
		&summary.Group.CreatedAt, &summary.Group.UpdatedAt, &summary.Group.Version, &summary.Group.ReadPolicy, &summary.Group.PostMessagePolicy, &summary.Group.PostThreadPolicy, &summary.Group.ModeratePolicy, &summary.LeaderboardScore, &summary.UnreadMessages,
		&lastID, &lastAuthorID, &lastName, &lastFigure, &lastBody, &lastState, &lastCreated)
	if err != nil {
		return nil, 0, err
	}
	summary.Group.ForumEnabled = true
	if lastID != nil {
		summary.LastPost = &grouprecord.Post{ID: *lastID, GroupID: summary.Group.ID, AuthorID: *lastAuthorID, AuthorName: *lastName, AuthorFigure: *lastFigure, Body: *lastBody, State: *lastState, CreatedAt: *lastCreated}
	}
	return []grouprecord.ForumSummary{summary}, 1, nil
}

// Threads lists one bounded forum thread page.
func (repository *Repository) Threads(ctx context.Context, playerID int64, groupID int64, offset int, limit int, includeHidden bool) ([]grouprecord.Thread, int32, error) {
	hidden := ` and state in (0,1)`
	if includeHidden {
		hidden = ``
	}
	var total int32
	if err := repository.executor(ctx).QueryRow(ctx, `select count(*) from social_group_forum_threads where group_id=$1`+hidden, groupID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := repository.executor(ctx).Query(ctx, `select thread.id,thread.group_id,thread.author_player_id,thread.author_name,thread.subject,thread.state,thread.pinned,thread.locked,thread.post_count,coalesce((select count(*) from social_group_forum_posts unread where unread.thread_id=thread.id and unread.state=0 and unread.id>coalesce(marker.last_message_id,0)),0),thread.last_post_id,thread.last_author_player_id,thread.last_author_name,thread.last_posted_at,thread.moderator_player_id,thread.moderator_name,thread.moderation_reason,thread.moderated_at,thread.created_at,thread.updated_at,thread.version from social_group_forum_threads thread left join social_group_forum_read_markers marker on marker.group_id=thread.group_id and marker.player_id=$4 where thread.group_id=$1`+hidden+` order by thread.pinned desc,thread.last_posted_at desc,thread.id desc offset $2 limit $3`, groupID, offset, limit, playerID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	threads := make([]grouprecord.Thread, 0, limit)
	for rows.Next() {
		thread, scanErr := scanThread(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		threads = append(threads, thread)
	}
	if _, err = repository.executor(ctx).Exec(ctx, `insert into social_group_forum_views(group_id,player_id) values($1,$2) on conflict(group_id,player_id,viewed_on) do update set view_count=least(social_group_forum_views.view_count+1,100)`, groupID, playerID); err != nil {
		return nil, 0, err
	}
	return threads, total, rows.Err()
}

// Thread returns one forum thread.
func (repository *Repository) Thread(ctx context.Context, groupID int64, threadID int64, includeHidden bool) (grouprecord.Thread, bool, error) {
	hidden := ` and state in (0,1)`
	if includeHidden {
		hidden = ``
	}
	thread, err := scanThread(repository.executor(ctx).QueryRow(ctx, `select id,group_id,author_player_id,author_name,subject,state,pinned,locked,post_count,0,last_post_id,last_author_player_id,last_author_name,last_posted_at,moderator_player_id,moderator_name,moderation_reason,moderated_at,created_at,updated_at,version from social_group_forum_threads where group_id=$1 and id=$2`+hidden, groupID, threadID))
	if errors.Is(err, pgx.ErrNoRows) {
		return grouprecord.Thread{}, false, nil
	}
	return thread, err == nil, err
}

// Posts lists one bounded thread message page.
func (repository *Repository) Posts(ctx context.Context, playerID int64, groupID int64, threadID int64, offset int, limit int, includeHidden bool) ([]grouprecord.Post, int32, error) {
	hidden := ` and post.state=0`
	if includeHidden {
		hidden = ``
	}
	var total int32
	if err := repository.executor(ctx).QueryRow(ctx, `select count(*) from social_group_forum_posts post where post.group_id=$1 and post.thread_id=$2`+hidden, groupID, threadID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := repository.executor(ctx).Query(ctx, `select post.id,post.group_id,post.thread_id,post.ordinal,post.author_player_id,post.author_name,post.author_figure,post.body,post.state,post.moderator_player_id,post.moderator_name,post.moderation_reason,post.moderated_at,(select count(*) from social_group_forum_posts authored where authored.group_id=post.group_id and authored.author_player_id=post.author_player_id),post.created_at,post.updated_at,post.version from social_group_forum_posts post where post.group_id=$1 and post.thread_id=$2`+hidden+` order by post.ordinal offset $3 limit $4`, groupID, threadID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	posts := make([]grouprecord.Post, 0, limit)
	for rows.Next() {
		post, scanErr := scanPost(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		posts = append(posts, post)
	}
	return posts, total, rows.Err()
}

// UnreadCount returns authorized hotel-wide forum unread total.
func (repository *Repository) UnreadCount(ctx context.Context, playerID int64, staff bool) (int32, error) {
	var count int32
	err := repository.executor(ctx).QueryRow(ctx, `select count(*) from social_group_forum_posts post join social_groups groups on groups.id=post.group_id and groups.deactivated_at is null and groups.forum_enabled left join social_group_members membership on membership.group_id=groups.id and membership.player_id=$1 left join social_group_forum_read_markers marker on marker.group_id=groups.id and marker.player_id=$1 where post.state=0 and post.id>coalesce(marker.last_message_id,0) and ($2 or groups.forum_read_policy=0 or (membership.player_id is not null and (groups.forum_read_policy=1 or (groups.forum_read_policy=2 and membership.role<=1) or (groups.forum_read_policy=3 and membership.role=0))))`, playerID, staff).Scan(&count)
	return count, err
}
