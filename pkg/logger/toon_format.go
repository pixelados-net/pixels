package logger

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"go.uber.org/zap/buffer"
)

var (
	// keyPattern detects TOON keys that do not need quoting.
	keyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.]*$`)

	// numericLikePattern detects strings that TOON must quote.
	numericLikePattern = regexp.MustCompile(`^-?\d+(?:\.\d+)?(?:e[+-]?\d+)?$`)

	// toonFieldOrder defines stable primary log field order.
	toonFieldOrder = []string{"lvl", "msg", "error", "cid", "state", "header", "bytes", "payload"}
)

// toonField stores one ordered TOON object field.
type toonField struct {
	// key stores the encoded object key.
	key string
	// value stores the JSON-like value.
	value interface{}
}

// encodeToonFields writes ordered fields as one TOON document.
func encodeToonFields(fields []toonField) *buffer.Buffer {
	output := buffer.NewPool().Get()
	for index, field := range fields {
		if index > 0 {
			output.AppendString(", ")
		}
		writeInlineToonField(output, field)
	}

	output.AppendByte('\n')

	return output
}

// writeInlineToonField writes one inline TOON object field.
func writeInlineToonField(output *buffer.Buffer, field toonField) {
	output.AppendString(encodeToonKey(field.key))
	output.AppendString(": ")
	output.AppendString(encodeInlineToonValue(field.value))
}

// encodeInlineToonValue encodes one inline TOON field value.
func encodeInlineToonValue(value interface{}) string {
	switch typed := value.(type) {
	case map[string]interface{}:
		return encodeInlineToonMap(typed)
	case []interface{}:
		return encodeInlineToonArray(typed)
	default:
		return encodeToonValue(typed)
	}
}

// encodeInlineToonMap encodes one nested object inline.
func encodeInlineToonMap(values map[string]interface{}) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder strings.Builder
	builder.WriteByte('{')
	for _, key := range keys {
		if builder.Len() > 1 {
			builder.WriteString(", ")
		}
		builder.WriteString(encodeToonKey(key))
		builder.WriteString(": ")
		builder.WriteString(encodeInlineToonValue(values[key]))
	}
	builder.WriteByte('}')

	return builder.String()
}

// encodeInlineToonArray encodes one inline array.
func encodeInlineToonArray(values []interface{}) string {
	var builder strings.Builder
	builder.WriteByte('[')
	builder.WriteString(strconv.Itoa(len(values)))
	builder.WriteString("]: ")
	for index, value := range values {
		if index > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(encodeInlineToonValue(value))
	}
	return builder.String()
}

// encodeToonKey encodes an object key.
func encodeToonKey(value string) string {
	if keyPattern.MatchString(value) {
		return value
	}

	return quoteToonString(value)
}

// encodeToonValue encodes one primitive TOON value.
func encodeToonValue(value interface{}) string {
	switch typed := value.(type) {
	case nil:
		return "null"
	case bool:
		return strconv.FormatBool(typed)
	case string:
		return encodeToonString(typed)
	case float64:
		return strconv.FormatFloat(typed, 'g', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(typed), 'g', -1, 32)
	default:
		return encodeToonString(fmt.Sprint(typed))
	}
}

// encodeToonString encodes one TOON string value.
func encodeToonString(value string) string {
	if shouldQuoteToonString(value) {
		return quoteToonString(value)
	}

	return value
}

// shouldQuoteToonString reports whether a TOON string requires quotes.
func shouldQuoteToonString(value string) bool {
	if value == "" || strings.TrimSpace(value) != value {
		return true
	}
	if value == "true" || value == "false" || value == "null" || value == "-" || strings.HasPrefix(value, "-") {
		return true
	}
	if numericLikePattern.MatchString(value) {
		return true
	}

	return strings.ContainsAny(value, ":\"\\[]{},") || hasToonControl(value)
}

// hasToonControl reports whether a string contains control bytes.
func hasToonControl(value string) bool {
	for _, char := range value {
		if char >= 0 && char < 0x20 {
			return true
		}
	}

	return false
}

// quoteToonString returns a quoted TOON string.
func quoteToonString(value string) string {
	return strconv.Quote(value)
}

// normalizeToonField rewrites known verbose log fields for TOON.
func normalizeToonField(key string, value interface{}) (toonField, bool) {
	switch key {
	case "level":
		return toonField{key: "lvl", value: value}, true
	case "connection_id":
		return toonField{key: "cid", value: shortToonValue(toonStringValue(value))}, true
	case "connection_kind":
		return toonField{}, false
	case "packet_header":
		return toonField{key: "header", value: value}, true
	case "packet_payload_size":
		return toonField{key: "bytes", value: value}, true
	case "packet_payload":
		return toonField{key: "payload", value: value}, true
	case "disconnect_message":
		return toonField{key: "error", value: value}, true
	default:
		return toonField{key: key, value: value}, true
	}
}

// toonStringValue returns a string representation for field normalization.
func toonStringValue(value interface{}) string {
	if text, ok := value.(string); ok {
		return text
	}

	return encodeToonValue(value)
}

// shortToonValue returns the compact visible prefix for a log value.
func shortToonValue(value string) string {
	if len(value) <= toonConnectionIDSize {
		return value
	}

	return value[:toonConnectionIDSize]
}
