// Package behavior owns immutable pet command definitions and chat execution.
package behavior

import (
	"fmt"
	"strings"
	"time"

	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
)

// Definition stores one pet speech command action.
type Definition struct {
	// ID identifies the protocol command.
	ID int32
	// Keyword stores the normalized speech keyword.
	Keyword string
	// Aliases stores additional normalized speech forms.
	Aliases []string
	// Action stores runtime behavior.
	Action petruntime.CommandAction
}

// Registry stores fixed command definitions by ID and keyword.
type Registry struct {
	// byID stores definitions without map lookup in command execution.
	byID [47]Definition
	// present reports registered command slots.
	present [47]bool
	// byKeyword resolves chat input outside the room tick.
	byKeyword map[string]int32
}

// NewRegistry creates and validates all 46 protocol commands.
func NewRegistry() *Registry {
	registry := &Registry{byKeyword: make(map[string]int32, 60)}
	for _, definition := range builtinDefinitions() {
		registry.byID[definition.ID] = definition
		registry.present[definition.ID] = true
		registry.byKeyword[definition.Keyword] = definition.ID
		for _, alias := range definition.Aliases {
			registry.byKeyword[alias] = definition.ID
		}
	}
	return registry
}

// Find returns one command definition by protocol identifier.
func (registry *Registry) Find(id int32) (Definition, bool) {
	if registry == nil || id < 0 || id >= int32(len(registry.byID)) || !registry.present[id] {
		return Definition{}, false
	}
	return registry.byID[id], true
}

// Resolve returns one command definition by normalized keyword.
func (registry *Registry) Resolve(keyword string) (Definition, bool) {
	id, found := registry.byKeyword[strings.ToLower(strings.TrimSpace(keyword))]
	if !found {
		return Definition{}, false
	}
	return registry.Find(id)
}

// Validate confirms every canonical command slot exists except historical gap 39.
func (registry *Registry) Validate() error {
	for id := int32(0); id <= 46; id++ {
		if id == 39 {
			continue
		}
		if _, found := registry.Find(id); !found {
			return fmt.Errorf("pet command %d is not registered", id)
		}
	}
	return nil
}

// action creates one temporary status action.
func action(id int32, keyword string, status string) Definition {
	return Definition{ID: id, Keyword: keyword, Action: petruntime.CommandAction{ID: id, Mode: petruntime.ActionStatus, StatusKey: status, Duration: 2 * time.Second}}
}

// mode creates one control action.
func mode(id int32, keyword string, value petruntime.ActionMode) Definition {
	return Definition{ID: id, Keyword: keyword, Action: petruntime.CommandAction{ID: id, Mode: value}}
}

// need creates one contextual product action.
func need(id int32, keyword string, value petruntime.CommandNeed) Definition {
	return Definition{ID: id, Keyword: keyword, Action: petruntime.CommandAction{ID: id, Mode: petruntime.ActionNeed, Need: value}}
}
