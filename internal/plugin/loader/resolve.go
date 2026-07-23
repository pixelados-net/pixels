package loader

import (
	"fmt"
	"sort"
)

// resolve rejects duplicates and returns a dependency-first topological order.
func resolve(candidates []candidate) ([]candidate, map[string]error) {
	blocked := make(map[string]error)
	grouped := make(map[string][]candidate, len(candidates))
	for _, current := range candidates {
		grouped[current.metadata.Name] = append(grouped[current.metadata.Name], current)
	}
	available := make(map[string]candidate, len(grouped))
	for name, group := range grouped {
		if len(group) != 1 {
			blocked[name] = fmt.Errorf("%w: duplicate name", ErrInvalidMetadata)
			continue
		}
		available[name] = group[0]
	}
	names := make([]string, 0, len(available))
	for name := range available {
		names = append(names, name)
	}
	sort.Strings(names)
	state := make(map[string]uint8, len(available))
	ordered := make([]candidate, 0, len(available))
	var visit func(string) error
	visit = func(name string) error {
		switch state[name] {
		case 1:
			return fmt.Errorf("%w: cycle at %s", ErrDependency, name)
		case 2:
			return blocked[name]
		}
		current, found := available[name]
		if !found {
			return fmt.Errorf("%w: missing %s", ErrDependency, name)
		}
		state[name] = 1
		for _, dependency := range current.metadata.Dependencies {
			if err := visit(dependency); err != nil {
				blocked[name] = err
				state[name] = 2
				return err
			}
		}
		state[name] = 2
		ordered = append(ordered, current)

		return nil
	}
	for _, name := range names {
		_ = visit(name)
	}
	filtered := ordered[:0]
	for _, current := range ordered {
		if blocked[current.metadata.Name] == nil {
			filtered = append(filtered, current)
		}
	}

	return filtered, blocked
}
