package logger

import (
	"encoding/json"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

const (
	// FormatTOON names the registered zap TOON encoder.
	FormatTOON = "toon"

	// toonConnectionIDSize is the visible connection identifier prefix length.
	toonConnectionIDSize = 8
)

func init() {
	_ = zap.RegisterEncoder(FormatTOON, func(config zapcore.EncoderConfig) (zapcore.Encoder, error) {
		return newToonEncoder(config), nil
	})
}

// toonEncoder converts zap JSON entries into TOON documents.
type toonEncoder struct {
	zapcore.Encoder
}

// newToonEncoder creates a TOON zap encoder.
func newToonEncoder(config zapcore.EncoderConfig) zapcore.Encoder {
	return &toonEncoder{Encoder: zapcore.NewJSONEncoder(config)}
}

// Clone copies the encoder.
func (encoder *toonEncoder) Clone() zapcore.Encoder {
	return &toonEncoder{Encoder: encoder.Encoder.Clone()}
}

// EncodeEntry encodes one log entry as TOON.
func (encoder *toonEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	data, err := encoder.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, err
	}
	defer data.Free()

	ordered, err := decodeToonLog(data.String())
	if err != nil {
		return nil, err
	}

	return encodeToonFields(ordered), nil
}

// decodeToonLog decodes zap JSON into ordered TOON fields.
func decodeToonLog(line string) ([]toonField, error) {
	var values map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &values); err != nil {
		return nil, err
	}

	fields := make([]toonField, 0, len(values))
	for _, key := range toonFieldOrder {
		field, ok := toonFieldFromMap(values, key)
		if ok {
			fields = append(fields, field)
		}
	}
	for key, value := range values {
		if toonOrderedKey(key) {
			continue
		}
		field, keep := normalizeToonField(key, value)
		if keep {
			fields = append(fields, field)
		}
	}

	return fields, nil
}

// toonFieldFromMap extracts and normalizes one field.
func toonFieldFromMap(values map[string]interface{}, key string) (toonField, bool) {
	value, ok := values[key]
	if !ok {
		return toonField{}, false
	}

	return normalizeToonField(key, value)
}

// toonOrderedKey reports whether a key was already handled.
func toonOrderedKey(key string) bool {
	for _, ordered := range toonFieldOrder {
		if key == ordered {
			return true
		}
	}

	return false
}

// toonLevelEncoder writes compact numeric level identifiers.
func toonLevelEncoder(level zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
	switch level {
	case zapcore.DebugLevel:
		encoder.AppendInt(0)
	case zapcore.InfoLevel:
		encoder.AppendInt(1)
	case zapcore.WarnLevel:
		encoder.AppendInt(2)
	case zapcore.ErrorLevel:
		encoder.AppendInt(3)
	case zapcore.DPanicLevel:
		encoder.AppendInt(4)
	case zapcore.PanicLevel:
		encoder.AppendInt(5)
	case zapcore.FatalLevel:
		encoder.AppendInt(6)
	default:
		encoder.AppendInt(int(level))
	}
}
