package configuration

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
)

var (
	// ErrUnsupported reports an interaction absent from the canonical registry.
	ErrUnsupported = errors.New("unsupported WIRED interaction")
	// ErrInvalid reports descriptor settings that violate its schema.
	ErrInvalid = errors.New("invalid WIRED settings")
)

// Compiler validates records and creates immutable runtime generations.
type Compiler struct {
	// registry resolves canonical behavior descriptors.
	registry *registry.Registry
	// config stores validated execution limits.
	config roomwired.Config
}

// NewCompiler creates a WIRED compiler.
func NewCompiler(registered *registry.Registry, config roomwired.Config) *Compiler {
	return &Compiler{registry: registered, config: config.Normalize()}
}

// Compile creates one immutable room generation.
func (compiler *Compiler) Compile(roomID int64, generationID uint64, records []record.Config) (*Generation, error) {
	generation := &Generation{ID: generationID, RoomID: roomID, Nodes: make(map[int64]*Node, len(records)), Stacks: make(map[Point]*Stack)}
	for _, stored := range records {
		if stored.RoomID != roomID {
			return nil, fmt.Errorf("%w: cross-room node %d", ErrInvalid, stored.ItemID)
		}
		node, err := compiler.CompileNode(stored)
		if err != nil {
			return nil, err
		}
		generation.Nodes[node.ItemID] = node
		stack := generation.Stacks[node.Point]
		if stack == nil {
			stack = &Stack{Point: node.Point}
			generation.Stacks[node.Point] = stack
		}
		switch node.Descriptor.Family {
		case registry.FamilyTrigger:
			stack.Triggers = append(stack.Triggers, node)
			generation.Triggers = append(generation.Triggers, node)
		case registry.FamilyCondition:
			stack.Conditions = append(stack.Conditions, node)
		case registry.FamilyEffect:
			stack.Effects = append(stack.Effects, node)
		case registry.FamilyExtra:
			stack.Extras = append(stack.Extras, node)
			applyExtra(stack, node.Descriptor.Key)
		}
	}
	for _, stack := range generation.Stacks {
		if stack.Unseen {
			stack.Random = false
		}
		sortNodes(stack.Triggers)
		sortNodes(stack.Conditions)
		sortNodes(stack.Effects)
		sortNodes(stack.Extras)
	}
	sortNodes(generation.Triggers)
	return generation, nil
}

// CompileNode validates and compiles one durable node.
func (compiler *Compiler) CompileNode(stored record.Config) (*Node, error) {
	descriptor, found := compiler.registry.Resolve(stored.Interaction)
	if !found {
		return nil, fmt.Errorf("%w: %s", ErrUnsupported, stored.Interaction)
	}
	if err := compiler.validate(stored, descriptor); err != nil {
		return nil, err
	}
	parameters, err := compileParameters(stored, descriptor)
	if err != nil {
		return nil, err
	}
	targets := append([]record.Target(nil), stored.Targets...)
	values := append([]int32(nil), stored.IntParams...)
	parameters.Values = values
	return &Node{
		ItemID: stored.ItemID, RoomID: stored.RoomID, SpriteID: stored.SpriteID,
		Point: Point{X: stored.X, Y: stored.Y}, Descriptor: descriptor,
		Parameters: parameters, SelectionMode: stored.SelectionMode,
		Delay:   time.Duration(stored.DelayPulses) * 500 * time.Millisecond,
		Version: stored.Version, Targets: targets,
	}, nil
}

// applyExtra applies one stack-level add-on.
func applyExtra(stack *Stack, key string) {
	switch key {
	case "wf_xtra_random":
		stack.Random = true
	case "wf_xtra_unseen":
		stack.Unseen = true
	case "wf_xtra_or_eval":
		stack.Or = true
	}
}

// sortNodes stabilizes execution independently of database iteration order.
func sortNodes(nodes []*Node) {
	sort.Slice(nodes, func(left int, right int) bool { return nodes[left].ItemID < nodes[right].ItemID })
}

// compileParameters parses strings and durations outside execution hot paths.
func compileParameters(stored record.Config, descriptor registry.Descriptor) (Parameters, error) {
	parameters := Parameters{Text: strings.TrimSpace(stored.StringParam)}
	if descriptor.Key == "wf_cnd_wearing_badge" || descriptor.Key == "wf_cnd_not_wearing_b" {
		parameters.Text = strings.ToUpper(parameters.Text)
	}
	if strings.HasPrefix(descriptor.Key, "wf_act_bot_") || strings.HasPrefix(descriptor.Key, "wf_trg_bot_") {
		parts := strings.SplitN(stored.StringParam, "\t", 2)
		parameters.Name = strings.TrimSpace(parts[0])
		if len(parts) == 2 {
			parameters.Message = strings.TrimSpace(parts[1])
		}
	}
	if descriptor.Key == "wf_act_give_respect" || descriptor.Key == "wf_act_give_handitem" || descriptor.Key == "wf_act_give_effect" {
		parsed, err := strconv.ParseInt(parameters.Text, 10, 32)
		if err != nil || parsed < 0 {
			return Parameters{}, fmt.Errorf("%w: numeric compatibility value", ErrInvalid)
		}
		parameters.Number = int32(parsed)
	}
	if descriptor.Key == "wf_trg_period_long" || descriptor.Key == "wf_trg_at_time_long" {
		if len(stored.IntParams) > 0 {
			parameters.Duration = time.Duration(stored.IntParams[0]) * 5 * time.Second
		}
	} else if strings.Contains(descriptor.Key, "period") || strings.Contains(descriptor.Key, "given_time") || strings.Contains(descriptor.Key, "time_more") || strings.Contains(descriptor.Key, "time_less") {
		if len(stored.IntParams) > 0 {
			parameters.Duration = time.Duration(stored.IntParams[0]) * 500 * time.Millisecond
		}
	}
	return parameters, nil
}
