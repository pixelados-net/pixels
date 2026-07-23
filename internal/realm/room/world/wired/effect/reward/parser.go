// Package reward owns normalized WIRED reward parsing and delivery.
package reward

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
)

// Parse converts Nitro's compact reward editor string into normalized rows.
func Parse(value string) ([]record.Reward, error) {
	parts := strings.Split(value, ";")
	rewards := make([]record.Reward, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.Split(part, ",")
		if len(fields) != 3 {
			return nil, fmt.Errorf("invalid reward entry")
		}
		weight, err := strconv.ParseInt(strings.TrimSpace(fields[2]), 10, 32)
		if err != nil || weight <= 0 || weight > 1000000 {
			return nil, fmt.Errorf("invalid reward weight")
		}
		kind, reference, amount, err := parseReference(fields[0], fields[1])
		if err != nil {
			return nil, err
		}
		rewards = append(rewards, record.Reward{Ordinal: len(rewards), Kind: kind, Reference: reference, Amount: amount, Weight: int32(weight)})
	}
	if len(rewards) == 0 || len(rewards) > 100 {
		return nil, fmt.Errorf("invalid reward count")
	}
	return rewards, nil
}

// parseReference maps badge and Arcturus-compatible item codes.
func parseReference(flag string, value string) (string, string, int64, error) {
	value = strings.TrimSpace(value)
	if strings.TrimSpace(flag) == "0" {
		if value == "" || len(value) > 64 {
			return "", "", 0, fmt.Errorf("invalid badge reward")
		}
		return "badge", strings.ToUpper(value), 1, nil
	}
	parts := strings.SplitN(value, "#", 2)
	if len(parts) == 1 {
		parts = []string{"furniture", value}
	}
	kind := strings.ToLower(strings.TrimSpace(parts[0]))
	reference := strings.TrimSpace(parts[1])
	amount := int64(1)
	switch kind {
	case "furni":
		kind = "furniture"
	case "cata":
		kind = "catalog_offer"
	case "credits", "respect":
		parsed, err := strconv.ParseInt(reference, 10, 64)
		if err != nil || parsed <= 0 {
			return "", "", 0, fmt.Errorf("invalid reward amount")
		}
		amount, reference = parsed, kind
	case "pixels":
		parsed, err := strconv.ParseInt(reference, 10, 64)
		if err != nil || parsed <= 0 {
			return "", "", 0, fmt.Errorf("invalid reward amount")
		}
		kind, reference, amount = "currency", "5", parsed
	default:
		if strings.HasPrefix(kind, "points") {
			currency := strings.TrimPrefix(kind, "points")
			if currency == "" {
				currency = "5"
			}
			parsed, err := strconv.ParseInt(reference, 10, 64)
			if err != nil || parsed <= 0 {
				return "", "", 0, fmt.Errorf("invalid currency reward")
			}
			kind, reference, amount = "currency", currency, parsed
		}
	}
	if kind != "furniture" && kind != "catalog_offer" && kind != "credits" && kind != "currency" && kind != "respect" {
		return "", "", 0, fmt.Errorf("unsupported reward kind")
	}
	if _, err := strconv.ParseInt(reference, 10, 64); err != nil && kind != "credits" && kind != "respect" {
		return "", "", 0, fmt.Errorf("invalid reward reference")
	}
	return kind, reference, amount, nil
}
