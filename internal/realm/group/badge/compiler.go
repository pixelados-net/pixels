package badge

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	groupobservability "github.com/niflaot/pixels/internal/realm/group/observability"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// ErrInvalidParts reports badge input outside the enabled registry.
var ErrInvalidParts = errors.New("invalid social group badge parts")

// Compiler validates normalized parts against one registry.
type Compiler struct {
	// registry stores immutable reference data.
	registry *Registry
	// metrics stores bounded process-wide group telemetry.
	metrics *groupobservability.Metrics
}

// NewCompiler creates a badge compiler.
func NewCompiler(registry *Registry) *Compiler { return &Compiler{registry: registry} }

// Compile validates and converts up to five badge layers to renderer code.
func (compiler *Compiler) Compile(parts []grouprecord.BadgePart) (code string, normalized []grouprecord.BadgePart, err error) {
	defer func() {
		result := groupobservability.Success
		if err != nil {
			result = groupobservability.Rejected
		}
		compiler.metrics.Record(groupobservability.BadgeCompile, groupobservability.KindDefault, result)
	}()
	snapshot, found := compiler.registry.Snapshot()
	if !found || len(parts) == 0 || len(parts) > 5 {
		return "", nil, ErrInvalidParts
	}
	normalized = make([]grouprecord.BadgePart, len(parts))
	var builder strings.Builder
	builder.Grow(len(parts) * 7)
	seen := make(map[int32]struct{}, len(parts))
	for index, part := range parts {
		kind := grouprecord.BadgeSymbol
		prefix := byte('s')
		if index == 0 {
			kind, prefix = grouprecord.BadgeBase, 'b'
		}
		if part.Kind != kind || part.Ordinal != int16(index) || part.ElementID <= 0 || part.ColorID <= 0 || part.Position < 0 || part.Position > 9 {
			return "", nil, ErrInvalidParts
		}
		if !snapshot.Element(kind, part.ElementID) {
			return "", nil, ErrInvalidParts
		}
		family := grouprecord.SymbolColor
		if kind == grouprecord.BadgeBase {
			family = grouprecord.BaseColor
		}
		if _, ok := snapshot.Color(family, part.ColorID); !ok {
			return "", nil, ErrInvalidParts
		}
		if _, duplicate := seen[part.ElementID]; duplicate {
			return "", nil, ErrInvalidParts
		}
		seen[part.ElementID] = struct{}{}
		normalized[index] = part
		builder.WriteByte(prefix)
		builder.WriteString(fmt.Sprintf("%03d%02d%d", part.ElementID, part.ColorID, part.Position))
	}
	return builder.String(), normalized, nil
}

// SetMetrics attaches process-wide telemetry before serving requests.
func (compiler *Compiler) SetMetrics(metrics *groupobservability.Metrics) { compiler.metrics = metrics }

// Parse converts a renderer badge code to normalized layers.
func Parse(code string) ([]grouprecord.BadgePart, error) {
	if code == "" {
		return nil, ErrInvalidParts
	}
	parts := make([]grouprecord.BadgePart, 0, 5)
	for offset := 0; offset < len(code); {
		if len(parts) == 5 || offset+7 > len(code) {
			return nil, ErrInvalidParts
		}
		end := len(code)
		for index := offset + 7; index < len(code); index++ {
			if code[index] == 'b' || code[index] == 's' {
				end = index
				break
			}
		}
		if end-offset < 7 || end-offset > 8 {
			return nil, ErrInvalidParts
		}
		prefix := code[offset]
		kind := grouprecord.BadgeSymbol
		if len(parts) == 0 && prefix == 'b' {
			kind = grouprecord.BadgeBase
		} else if prefix != 's' {
			return nil, ErrInvalidParts
		}
		element, err := strconv.ParseInt(code[offset+1:offset+4], 10, 32)
		if err != nil {
			return nil, ErrInvalidParts
		}
		color, err := strconv.ParseInt(code[offset+4:end-1], 10, 32)
		if err != nil {
			return nil, ErrInvalidParts
		}
		position, err := strconv.ParseInt(code[end-1:end], 10, 32)
		if err != nil {
			return nil, ErrInvalidParts
		}
		parts = append(parts, grouprecord.BadgePart{Ordinal: int16(len(parts)), Kind: kind, ElementID: int32(element), ColorID: int32(color), Position: int32(position)})
		offset = end
	}
	return parts, nil
}
