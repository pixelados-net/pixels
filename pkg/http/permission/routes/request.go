package routes

import (
	"net/url"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
)

// GroupRequest contains permission group creation fields.
type GroupRequest struct {
	// Name stores the unique group name.
	Name string `json:"name"`
	// Weight stores group resolution priority.
	Weight int32 `json:"weight"`
	// Prefix stores a future chat prefix.
	Prefix string `json:"prefix"`
	// PrefixColor stores a future chat prefix color.
	PrefixColor string `json:"prefixColor"`
	// ParentGroupID identifies the optional inherited group.
	ParentGroupID *int64 `json:"parentGroupId"`
}

// GroupPatchRequest contains optional permission group fields.
type GroupPatchRequest struct {
	// Name replaces the group name.
	Name *string `json:"name"`
	// Weight replaces group resolution priority.
	Weight *int32 `json:"weight"`
	// Prefix replaces the future chat prefix.
	Prefix *string `json:"prefix"`
	// PrefixColor replaces the future chat prefix color.
	PrefixColor *string `json:"prefixColor"`
	// ParentGroupID replaces the inherited group.
	ParentGroupID *int64 `json:"parentGroupId"`
	// ClearParent removes the inherited group.
	ClearParent bool `json:"clearParent"`
}

// NodeRequest contains one permission grant mutation.
type NodeRequest struct {
	// Node identifies the capability or wildcard.
	Node permission.Node `json:"node"`
	// Allowed reports whether the grant allows or denies.
	Allowed bool `json:"allowed"`
}

// groupParams maps a group request into service input.
func groupParams(request GroupRequest) permissionservice.CreateGroupParams {
	return permissionservice.CreateGroupParams{Name: request.Name, Weight: request.Weight, Prefix: request.Prefix,
		PrefixColor: request.PrefixColor, ParentGroupID: request.ParentGroupID}
}

// groupPatch maps a group patch request into service input.
func groupPatch(request GroupPatchRequest) permissionservice.UpdateGroupParams {
	params := permissionservice.UpdateGroupParams{Name: request.Name, Weight: request.Weight,
		Prefix: request.Prefix, PrefixColor: request.PrefixColor}
	if request.ClearParent {
		var parent *int64
		params.ParentGroupID = &parent
	} else if request.ParentGroupID != nil {
		params.ParentGroupID = &request.ParentGroupID
	}

	return params
}

// routeID parses one positive route identifier.
func routeID(ctx *fiber.Ctx, name string) (int64, error) {
	value, err := strconv.ParseInt(ctx.Params(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid permission "+name)
	}

	return value, nil
}

// routeNode decodes one permission node route parameter.
func routeNode(ctx *fiber.Ctx) (permission.Node, error) {
	value, err := url.PathUnescape(ctx.Params("node"))
	if err != nil || value == "" {
		return "", fiber.NewError(fiber.StatusBadRequest, "invalid permission node")
	}

	return permission.Node(value), nil
}
