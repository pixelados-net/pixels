package routes

import (
	"github.com/gofiber/fiber/v2"
	grouppolicy "github.com/niflaot/pixels/internal/realm/group/policy"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// forumSettings returns forum entitlement, policies, counts, and version.
func (dependencies Dependencies) forumSettings(ctx *fiber.Ctx) error {
	actorID, err := dependencies.readAuthorized(ctx, grouppolicy.ForumManageAny)
	if err != nil {
		return err
	}
	groupID, err := positiveParam(ctx, "groupId")
	if err != nil {
		return err
	}
	summary, access, err := dependencies.Forum.Stats(ctx.Context(), actorID, groupID)
	if err != nil {
		return groupError(err)
	}
	return ctx.JSON(struct {
		Summary grouprecord.ForumSummary `json:"summary"`
		Access  any                      `json:"access"`
	}{Summary: summary, Access: access})
}

// updateForumSettings updates entitlement and all four policies optimistically.
func (dependencies Dependencies) updateForumSettings(ctx *fiber.Ctx) error {
	request := struct {
		VersionRequest
		Enabled     bool               `json:"enabled"`
		Read        grouprecord.Policy `json:"read"`
		PostMessage grouprecord.Policy `json:"postMessage"`
		PostThread  grouprecord.Policy `json:"postThread"`
		Moderate    grouprecord.Policy `json:"moderate"`
	}{}
	groupID, err := dependencies.mutation(ctx, "groupId", &request.AuditRequest, grouppolicy.ForumManageAny, &request)
	if err != nil {
		return err
	}
	group, err := dependencies.Forum.UpdateSettings(auditContext(ctx, request.AuditRequest), request.ActorPlayerID, groupID, request.Version, request.Enabled, request.Read, request.PostMessage, request.PostThread, request.Moderate)
	return groupMutation(ctx, group, err)
}

// threads returns one bounded forum thread page including hidden rows for staff.
func (dependencies Dependencies) threads(ctx *fiber.Ctx) error {
	actorID, err := dependencies.readAuthorized(ctx, grouppolicy.ForumModerateAny)
	if err != nil {
		return err
	}
	groupID, err := positiveParam(ctx, "groupId")
	if err != nil {
		return err
	}
	start, err := boundedInt(ctx.Query("start"), 0, 1000000)
	if err != nil {
		return err
	}
	amount, err := boundedInt(ctx.Query("amount"), 50, 50)
	if err != nil || amount < 1 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid forum page amount")
	}
	items, total, _, err := dependencies.Forum.Threads(ctx.Context(), actorID, groupID, start, amount)
	if err != nil {
		return groupError(err)
	}
	return ctx.JSON(struct {
		Items []grouprecord.Thread `json:"items"`
		Total int32                `json:"total"`
	}{Items: items, Total: total})
}

// thread returns one retained thread and bounded post page.
func (dependencies Dependencies) thread(ctx *fiber.Ctx) error {
	actorID, err := dependencies.readAuthorized(ctx, grouppolicy.ForumModerateAny)
	if err != nil {
		return err
	}
	groupID, err := positiveParam(ctx, "groupId")
	if err != nil {
		return err
	}
	threadID, err := positiveParam(ctx, "threadId")
	if err != nil {
		return err
	}
	start, err := boundedInt(ctx.Query("start"), 0, 1000000)
	if err != nil {
		return err
	}
	amount, err := boundedInt(ctx.Query("amount"), 50, 50)
	if err != nil || amount < 1 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid forum page amount")
	}
	value, _, err := dependencies.Forum.Thread(ctx.Context(), actorID, groupID, threadID)
	if err != nil {
		return groupError(err)
	}
	posts, total, _, err := dependencies.Forum.Posts(ctx.Context(), actorID, groupID, threadID, start, amount)
	if err != nil {
		return groupError(err)
	}
	return ctx.JSON(ForumThreadResponse{Thread: value, Posts: posts, Total: total})
}

// updateThread changes pin, lock, or retained moderation state.
func (dependencies Dependencies) updateThread(ctx *fiber.Ctx) error {
	request := struct {
		VersionRequest
		Pinned *bool                    `json:"pinned"`
		Locked *bool                    `json:"locked"`
		State  *grouprecord.ThreadState `json:"state"`
	}{}
	groupID, err := dependencies.mutation(ctx, "groupId", &request.AuditRequest, grouppolicy.ForumModerateAny, &request)
	if err != nil {
		return err
	}
	threadID, err := positiveParam(ctx, "threadId")
	if err != nil {
		return err
	}
	thread, err := dependencies.Forum.UpdateThread(auditContext(ctx, request.AuditRequest), request.ActorPlayerID, groupID, threadID, request.Version, request.Pinned, request.Locked, request.State, request.Reason)
	if err != nil {
		return groupError(err)
	}
	return ctx.JSON(thread)
}

// updatePost changes one retained message moderation state.
func (dependencies Dependencies) updatePost(ctx *fiber.Ctx) error {
	request := struct {
		VersionRequest
		State grouprecord.PostState `json:"state"`
	}{}
	groupID, err := dependencies.mutation(ctx, "groupId", &request.AuditRequest, grouppolicy.ForumModerateAny, &request)
	if err != nil {
		return err
	}
	postID, err := positiveParam(ctx, "postId")
	if err != nil {
		return err
	}
	post, err := dependencies.Forum.UpdatePost(auditContext(ctx, request.AuditRequest), request.ActorPlayerID, groupID, postID, request.Version, request.State, request.Reason)
	if err != nil {
		return groupError(err)
	}
	return ctx.JSON(post)
}
