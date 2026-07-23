// Package duration formats localized, human-readable durations.
package duration

import (
	"strconv"
	"time"

	"github.com/niflaot/pixels/pkg/i18n"
)

const (
	// secondOne identifies a singular second translation.
	secondOne i18n.Key = "duration.second.one"
	// secondOther identifies a plural second translation.
	secondOther i18n.Key = "duration.second.other"
	// minuteOne identifies a singular minute translation.
	minuteOne i18n.Key = "duration.minute.one"
	// minuteOther identifies a plural minute translation.
	minuteOther i18n.Key = "duration.minute.other"
	// hourOne identifies a singular hour translation.
	hourOne i18n.Key = "duration.hour.one"
	// hourOther identifies a plural hour translation.
	hourOther i18n.Key = "duration.hour.other"
	// dayOne identifies a singular day translation.
	dayOne i18n.Key = "duration.day.one"
	// dayOther identifies a plural day translation.
	dayOther i18n.Key = "duration.day.other"
)

// Format returns one rounded duration unit for a locale.
func Format(translator i18n.Translator, locale i18n.Locale, value time.Duration) string {
	count, one, other := selectUnit(value)
	key := other
	if count == 1 {
		key = one
	}
	if translator == nil {
		return ""
	}

	return translator.T(locale, key, i18n.Params{"count": strconv.FormatInt(count, 10)})
}

// Default returns one rounded duration unit for the default locale.
func Default(translator i18n.Translator, value time.Duration) string {
	count, one, other := selectUnit(value)
	key := other
	if count == 1 {
		key = one
	}
	if translator == nil {
		return ""
	}

	return translator.Default(key, i18n.Params{"count": strconv.FormatInt(count, 10)})
}

// DefaultParams returns the interpolation parameters for a default-locale duration.
func DefaultParams(translator i18n.Translator, value time.Duration) i18n.Params {
	return i18n.Params{"duration": Default(translator, value)}
}

// selectUnit selects the largest useful unit and rounds partial units upward.
func selectUnit(value time.Duration) (int64, i18n.Key, i18n.Key) {
	if value <= 0 {
		value = time.Second
	}
	switch {
	case value < time.Minute:
		return rounded(value, time.Second), secondOne, secondOther
	case value < time.Hour:
		return rounded(value, time.Minute), minuteOne, minuteOther
	case value < 24*time.Hour:
		return rounded(value, time.Hour), hourOne, hourOther
	default:
		return rounded(value, 24*time.Hour), dayOne, dayOther
	}
}

// rounded divides durations and rounds a partial unit upward without overflow.
func rounded(value time.Duration, unit time.Duration) int64 {
	count := int64(value / unit)
	if value%unit != 0 {
		count++
	}

	return max(1, count)
}
