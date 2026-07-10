package service

import (
	"context"
	"sort"

	"github.com/niflaot/pixels/internal/permission"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
)

// decision stores one internal permission resolution result.
type decision struct {
	// found reports whether a matching grant exists.
	found bool
	// allowed stores the matching grant decision.
	allowed bool
	// specificity stores matching node specificity.
	specificity int
	// depth stores inheritance distance from the selected group.
	depth int
	// source identifies the deciding holder.
	source string
}

// HasPermission reports whether a player currently holds one concrete node.
func (service *Service) HasPermission(ctx context.Context, playerID int64, node permission.Node) (bool, error) {
	resolved, err := service.resolve(ctx, playerID, node)
	return resolved.found && resolved.allowed, err
}

// resolve applies player, group weight, inheritance, and specificity precedence.
func (service *Service) resolve(ctx context.Context, playerID int64, node permission.Node) (decision, error) {
	if playerID <= 0 {
		return decision{}, ErrInvalidPlayerID
	}
	if !node.Concrete() {
		return decision{}, ErrInvalidNode
	}

	direct, err := service.playerNodes(ctx, playerID)
	if err != nil {
		return decision{}, err
	}
	if resolved := bestGrant(direct, node, 0, "player"); resolved.found {
		return resolved, nil
	}

	groups, err := service.playerGroups(ctx, playerID)
	if err != nil {
		return decision{}, err
	}
	for _, group := range groups {
		resolved, resolveErr := service.resolveGroup(ctx, group, node)
		if resolveErr != nil {
			return decision{}, resolveErr
		}
		if resolved.found {
			resolved.source = group.Name
			return resolved, nil
		}
	}

	return decision{}, nil
}

// resolveGroup resolves grants inside one group inheritance chain.
func (service *Service) resolveGroup(ctx context.Context, group permissionmodel.Group, node permission.Node) (decision, error) {
	var visited [16]int64
	var overflow map[int64]struct{}
	best := decision{specificity: -1}
	for depth := 0; ; depth++ {
		if depth < len(visited) {
			for index := 0; index < depth; index++ {
				if visited[index] == group.ID {
					return decision{}, ErrInheritanceCycle
				}
			}
			visited[depth] = group.ID
		} else {
			if overflow == nil {
				overflow = make(map[int64]struct{}, len(visited)+1)
				for _, groupID := range visited {
					overflow[groupID] = struct{}{}
				}
			}
			if _, found := overflow[group.ID]; found {
				return decision{}, ErrInheritanceCycle
			}
			overflow[group.ID] = struct{}{}
		}

		grants, err := service.groupNodes(ctx, group.ID)
		if err != nil {
			return decision{}, err
		}
		best = preferable(best, bestGrant(grants, node, depth, group.Name))
		if group.ParentGroupID == nil {
			return best, nil
		}
		parent, found, err := service.group(ctx, *group.ParentGroupID)
		if err != nil {
			return decision{}, err
		}
		if !found {
			return decision{}, ErrGroupNotFound
		}
		group = parent
	}
}

// bestGrant selects the most specific matching grant, preferring deny on ties.
func bestGrant(grants []permissionmodel.Grant, node permission.Node, depth int, source string) decision {
	best := decision{specificity: -1}
	for _, grant := range grants {
		specificity := grant.Node.Specificity(node)
		if specificity < 0 {
			continue
		}
		candidate := decision{found: true, allowed: grant.Allowed, specificity: specificity, depth: depth, source: source}
		best = preferable(best, candidate)
	}

	return best
}

// preferable selects specificity, child depth, and deny precedence in that order.
func preferable(left decision, right decision) decision {
	if !left.found {
		return right
	}
	if !right.found {
		return left
	}
	if left.specificity != right.specificity {
		if left.specificity > right.specificity {
			return left
		}
		return right
	}
	if left.depth != right.depth {
		if left.depth < right.depth {
			return left
		}
		return right
	}
	if !left.allowed {
		return left
	}

	return right
}

// playerNodes loads cached direct player grants.
func (service *Service) playerNodes(ctx context.Context, playerID int64) ([]permissionmodel.Grant, error) {
	return service.cache.PlayerNodes(ctx, playerID, func(ctx context.Context) ([]permissionmodel.Grant, error) {
		return service.store.ListPlayerNodes(ctx, playerID)
	})
}

// playerGroups loads cached player memberships in stable priority order.
func (service *Service) playerGroups(ctx context.Context, playerID int64) ([]permissionmodel.Group, error) {
	return service.cache.PlayerGroups(ctx, playerID, func(ctx context.Context) ([]permissionmodel.Group, error) {
		groups, err := service.store.ListGroupsByPlayer(ctx, playerID)
		if err != nil {
			return nil, err
		}
		sort.SliceStable(groups, func(left int, right int) bool {
			if groups[left].Weight == groups[right].Weight {
				return groups[left].ID < groups[right].ID
			}
			return groups[left].Weight > groups[right].Weight
		})

		return groups, nil
	})
}

// group loads one cached permission group.
func (service *Service) group(ctx context.Context, groupID int64) (permissionmodel.Group, bool, error) {
	return service.cache.Group(ctx, groupID, func(ctx context.Context) (permissionmodel.Group, bool, error) {
		return service.store.FindGroupByID(ctx, groupID)
	})
}

// groupNodes loads cached group grants.
func (service *Service) groupNodes(ctx context.Context, groupID int64) ([]permissionmodel.Grant, error) {
	return service.cache.GroupNodes(ctx, groupID, func(ctx context.Context) ([]permissionmodel.Grant, error) {
		return service.store.ListGroupNodes(ctx, groupID)
	})
}
