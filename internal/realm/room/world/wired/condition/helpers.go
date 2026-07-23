package condition

import "github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"

// everyTarget evaluates an all-target predicate.
func everyTarget(node *configuration.Node, predicate func(int64, int) (bool, error)) (Result, error) {
	if len(node.Targets) == 0 {
		return Result{}, nil
	}
	for index, target := range node.Targets {
		pass, err := predicate(target.ItemID, index)
		if err != nil {
			return Result{Valid: true}, err
		}
		if !pass {
			return Result{Valid: true}, nil
		}
	}
	return Result{Pass: true, Valid: true}, nil
}

// actorOnTargets evaluates event actor occupancy across selected furniture.
func actorOnTargets(node *configuration.Node, context Context, view View) (Result, error) {
	if context.Event.ActorID <= 0 {
		return Result{}, nil
	}
	for _, target := range node.Targets {
		pass, valid, err := view.ActorOn(context.Event, target.ItemID)
		if err != nil {
			return Result{Valid: valid}, err
		}
		if pass {
			return Result{Pass: true, Valid: valid}, nil
		}
		if !valid {
			return Result{}, nil
		}
	}
	return Result{Valid: true}, nil
}

// stackedTargets evaluates all or any base targets from the editor setting.
func stackedTargets(node *configuration.Node, view View) (Result, error) {
	any := len(node.Parameters.Values) > 0 && node.Parameters.Values[0] == 1
	matched := false
	for _, target := range node.Targets {
		pass, err := view.Stacked(target.ItemID)
		if err != nil {
			return Result{Valid: true}, err
		}
		matched = matched || pass
		if any && pass {
			return Result{Pass: true, Valid: true}, nil
		}
		if !any && !pass {
			return Result{Valid: true}, nil
		}
	}
	return Result{Pass: matched || len(node.Targets) > 0, Valid: len(node.Targets) > 0}, nil
}

// targetMatches applies Nitro's explicit ID, type, and context policies.
func targetMatches(node *configuration.Node, itemID int64, spriteID int32) bool {
	if node.SelectionMode == 0 {
		return false
	}
	for _, target := range node.Targets {
		if target.ItemID == itemID {
			return true
		}
		if node.SelectionMode >= 2 && spriteID > 0 && target.SpriteID == spriteID {
			return true
		}
	}
	return false
}

// first returns the first integer setting or zero.
func first(values []int32) int32 {
	if len(values) == 0 {
		return 0
	}
	return values[0]
}
