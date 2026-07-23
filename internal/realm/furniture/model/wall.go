package model

import "regexp"

// wallPositionPattern matches Nitro's modern wall coordinate notation.
var wallPositionPattern = regexp.MustCompile(`^:w=[0-9]{1,4},[0-9]{1,4} l=[0-9]{1,4},[0-9]{1,4} [lr]$`)

// ValidWallPosition reports whether a value uses Nitro's modern wall notation.
func ValidWallPosition(value string) bool {
	return wallPositionPattern.MatchString(value)
}
