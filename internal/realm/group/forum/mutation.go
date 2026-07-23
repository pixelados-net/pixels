package forum

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/niflaot/pixels/internal/permission"
	postedevent "github.com/niflaot/pixels/internal/realm/group/forum/events/posted"
	threadevent "github.com/niflaot/pixels/internal/realm/group/forum/events/threadchanged"
	updatedevent "github.com/niflaot/pixels/internal/realm/group/identity/events/updated"
	groupobservability "github.com/niflaot/pixels/internal/realm/group/observability"
	grouppolicy "github.com/niflaot/pixels/internal/realm/group/policy"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

func (service *Service) CreateThread(ctx context.Context, playerID int64, groupID int64, subject string, body string) (grouprecord.Thread, grouprecord.Post, error) {
	group, access, err := service.access(ctx, playerID, groupID)
	if err != nil || !allows(group.PostThreadPolicy, access) {
		return grouprecord.Thread{}, grouprecord.Post{}, grouprecord.ErrForbidden
	}
	subject, body, err = service.validateText(playerID, subject, body)
	if err != nil {
		return grouprecord.Thread{}, grouprecord.Post{}, err
	}
	player, found, err := service.players.FindByID(ctx, playerID)
	if err != nil || !found {
		return grouprecord.Thread{}, grouprecord.Post{}, grouprecord.ErrNotFound
	}
	thread, post, err := service.store.CreateThread(ctx, groupID, playerID, player.Player.Username, player.Profile.Look, subject, body)
	service.metrics.Record(groupobservability.ForumOperations, groupobservability.KindPost, forumMetricResult(err))
	if err == nil {
		service.publish(ctx, threadevent.Name, threadevent.Payload{GroupID: groupID, ThreadID: thread.ID, Version: thread.Version, Action: "create"})
		service.publish(ctx, postedevent.Name, postedevent.Payload{GroupID: groupID, ThreadID: thread.ID, PostID: post.ID, Version: post.Version})
	}
	return thread, post, err
}

// CreatePost creates one authorized reply.
func (service *Service) CreatePost(ctx context.Context, playerID int64, groupID int64, threadID int64, body string) (grouprecord.Post, error) {
	group, access, err := service.access(ctx, playerID, groupID)
	if err != nil || !allows(group.PostMessagePolicy, access) {
		return grouprecord.Post{}, grouprecord.ErrForbidden
	}
	thread, found, err := service.store.Thread(ctx, groupID, threadID, access.Staff)
	if err != nil || !found {
		return grouprecord.Post{}, grouprecord.ErrNotFound
	}
	if thread.Locked && !access.Staff {
		return grouprecord.Post{}, grouprecord.ErrClosed
	}
	_, body, err = service.validateText(playerID, "reply", body)
	if err != nil {
		return grouprecord.Post{}, err
	}
	player, found, err := service.players.FindByID(ctx, playerID)
	if err != nil || !found {
		return grouprecord.Post{}, grouprecord.ErrNotFound
	}
	post, err := service.store.CreatePost(ctx, groupID, threadID, playerID, player.Player.Username, player.Profile.Look, body)
	service.metrics.Record(groupobservability.ForumOperations, groupobservability.KindPost, forumMetricResult(err))
	if err == nil {
		service.publish(ctx, postedevent.Name, postedevent.Payload{GroupID: groupID, ThreadID: threadID, PostID: post.ID, Version: post.Version})
	}
	return post, err
}

// UpdateSettings changes forum entitlement and four policies.
func (service *Service) UpdateSettings(ctx context.Context, actorID int64, groupID int64, version int64, enabled bool, read grouprecord.Policy, postMessage grouprecord.Policy, postThread grouprecord.Policy, moderate grouprecord.Policy) (grouprecord.Group, error) {
	group, access, err := service.access(ctx, actorID, groupID)
	if err != nil {
		return grouprecord.Group{}, err
	}
	if access.Role != grouprecord.Owner && !access.Staff {
		allowed, checkErr := service.has(ctx, actorID, grouppolicy.ForumManageAny)
		if checkErr != nil || !allowed {
			return grouprecord.Group{}, grouprecord.ErrForbidden
		}
	}
	if !read.Valid() || !postMessage.Valid() || !postThread.Valid() || !moderate.Valid() {
		return grouprecord.Group{}, grouprecord.ErrInvalid
	}
	patch := grouprecord.GroupPatch{ForumEnabled: &enabled, ReadPolicy: &read, PostMessagePolicy: &postMessage, PostThreadPolicy: &postThread, ModeratePolicy: &moderate}
	updated, err := service.store.UpdateGroup(ctx, group.ID, version, patch)
	service.metrics.Record(groupobservability.ForumOperations, groupobservability.KindUpdate, forumMetricResult(err))
	if err == nil {
		service.publish(ctx, updatedevent.Name, updatedevent.Payload{GroupID: groupID, Version: updated.Version, Action: "forum_settings"})
	}
	return updated, err
}

// UpdateThread changes pin, lock, or moderation state.
func (service *Service) UpdateThread(ctx context.Context, actorID int64, groupID int64, threadID int64, version int64, pinned *bool, locked *bool, state *grouprecord.ThreadState, reason string) (grouprecord.Thread, error) {
	group, access, err := service.access(ctx, actorID, groupID)
	if err != nil || !allows(group.ModeratePolicy, access) {
		return grouprecord.Thread{}, grouprecord.ErrForbidden
	}
	if state != nil && !state.Valid() {
		return grouprecord.Thread{}, grouprecord.ErrInvalid
	}
	thread, err := service.store.UpdateThread(ctx, groupID, threadID, version, pinned, locked, state, actorID, strings.TrimSpace(reason))
	service.metrics.Record(groupobservability.ForumOperations, groupobservability.KindModerate, forumMetricResult(err))
	if err == nil {
		service.publish(ctx, threadevent.Name, threadevent.Payload{GroupID: groupID, ThreadID: thread.ID, Version: thread.Version, Action: "moderate"})
	}
	return thread, err
}

// UpdatePost changes one retained post moderation state.
func (service *Service) UpdatePost(ctx context.Context, actorID int64, groupID int64, postID int64, version int64, state grouprecord.PostState, reason string) (grouprecord.Post, error) {
	group, access, err := service.access(ctx, actorID, groupID)
	if err != nil || !allows(group.ModeratePolicy, access) {
		return grouprecord.Post{}, grouprecord.ErrForbidden
	}
	if !state.Valid() {
		return grouprecord.Post{}, grouprecord.ErrInvalid
	}
	post, err := service.store.UpdatePost(ctx, groupID, postID, version, state, actorID, strings.TrimSpace(reason))
	service.metrics.Record(groupobservability.ForumOperations, groupobservability.KindModerate, forumMetricResult(err))
	if err == nil {
		service.publish(ctx, postedevent.Name, postedevent.Payload{GroupID: groupID, ThreadID: post.ThreadID, PostID: post.ID, Version: post.Version})
	}
	return post, err
}

// UpdateReadMarkers advances a bounded collection monotonically.
func (service *Service) UpdateReadMarkers(ctx context.Context, playerID int64, markers []grouprecord.ReadMarker) ([]grouprecord.ReadMarker, error) {
	if len(markers) > service.config.ForumPageSize {
		return nil, grouprecord.ErrInvalid
	}
	result := make([]grouprecord.ReadMarker, 0, len(markers))
	for _, marker := range markers {
		marker.PlayerID = playerID
		group, access, err := service.access(ctx, playerID, marker.GroupID)
		if err != nil || !allows(group.ReadPolicy, access) || marker.LastMessageID < 0 {
			return nil, grouprecord.ErrForbidden
		}
		updated, err := service.store.UpdateReadMarker(ctx, marker)
		if err != nil {
			return nil, err
		}
		result = append(result, updated)
	}
	return result, nil
}

// UnreadCount returns the authorized hotel-wide unread total.
func (service *Service) UnreadCount(ctx context.Context, playerID int64) (int32, error) {
	staff, err := service.has(ctx, playerID, grouppolicy.ForumModerateAny)
	if err != nil {
		return 0, err
	}
	count, err := service.store.UnreadCount(ctx, playerID, staff)
	service.metrics.Record(groupobservability.ForumUnreadCache, groupobservability.KindDefault, forumMetricResult(err))
	return count, err
}

// readAccess validates page bounds and forum read policy.
func (service *Service) readAccess(ctx context.Context, playerID int64, groupID int64, start int, amount int) (grouprecord.Group, Access, error) {
	if start < 0 || amount < 1 || amount > service.config.ForumPageSize {
		return grouprecord.Group{}, Access{}, grouprecord.ErrInvalid
	}
	group, access, err := service.access(ctx, playerID, groupID)
	if err != nil {
		return grouprecord.Group{}, access, err
	}
	if !allows(group.ReadPolicy, access) {
		return grouprecord.Group{}, access, grouprecord.ErrForbidden
	}
	return group, access, nil
}

// access resolves one viewer's social role and hotel override.
func (service *Service) access(ctx context.Context, playerID int64, groupID int64) (grouprecord.Group, Access, error) {
	group, found, err := service.store.Group(ctx, groupID, false)
	if err != nil || !found || !group.ForumEnabled {
		return grouprecord.Group{}, Access{}, grouprecord.ErrNotFound
	}
	member, memberFound, err := service.store.Membership(ctx, groupID, playerID)
	if err != nil {
		return grouprecord.Group{}, Access{}, err
	}
	staff, err := service.has(ctx, playerID, grouppolicy.ForumModerateAny)
	return group, Access{Member: memberFound, Role: member.Role, Staff: staff}, err
}

// allows evaluates one policy without allocation.
func allows(policy grouprecord.Policy, access Access) bool {
	if access.Staff || policy == grouprecord.Everyone {
		return true
	}
	if !access.Member {
		return false
	}
	if policy == grouprecord.Members {
		return true
	}
	if policy == grouprecord.Admins {
		return access.Role <= grouprecord.Admin
	}
	return access.Role == grouprecord.Owner
}

// validateText bounds, filters, and rate-limits one write.
func (service *Service) validateText(playerID int64, subject string, body string) (string, string, error) {
	subject, body = strings.TrimSpace(subject), strings.TrimSpace(body)
	if utf8.RuneCountInString(subject) < 1 || utf8.RuneCountInString(subject) > service.config.ForumSubjectLimit || utf8.RuneCountInString(body) < 1 || utf8.RuneCountInString(body) > service.config.ForumMessageLimit {
		return "", "", grouprecord.ErrInvalid
	}
	service.rateMutex.Lock()
	now := time.Now()
	if len(service.lastPost) >= maxRateEntries {
		cutoff := now.Add(-service.config.ForumPostCooldown)
		for id, postedAt := range service.lastPost {
			if postedAt.Before(cutoff) {
				delete(service.lastPost, id)
			}
		}
		if len(service.lastPost) >= maxRateEntries {
			service.rateMutex.Unlock()
			return "", "", grouprecord.ErrLimit
		}
	}
	last := service.lastPost[playerID]
	if now.Sub(last) < service.config.ForumPostCooldown {
		service.rateMutex.Unlock()
		return "", "", grouprecord.ErrLimit
	}
	service.lastPost[playerID] = now
	service.rateMutex.Unlock()
	if service.filter != nil {
		subject, _ = service.filter.Censor(subject)
		body, _ = service.filter.Censor(body)
	}
	return subject, body, nil
}

// has resolves one optional hotel permission checker.
func (service *Service) has(ctx context.Context, playerID int64, node permission.Node) (bool, error) {
	if service.permissions == nil {
		return false, nil
	}
	return service.permissions.HasPermission(ctx, playerID, node)
}
