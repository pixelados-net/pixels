// Package permission contains typed permission identifiers and holder contracts.
package permission

import "strings"

const (
	// Wildcard grants every concrete node beneath its prefix.
	Wildcard = "*"
)

// Node identifies one dotted permission capability.
type Node string

// Valid reports whether a node uses valid dotted permission syntax.
func (node Node) Valid() bool {
	value := string(node)
	if value == Wildcard {
		return true
	}
	if value == "" || strings.HasPrefix(value, ".") || strings.HasSuffix(value, ".") {
		return false
	}

	segments := strings.Split(value, ".")
	for index, segment := range segments {
		if segment == "" || (segment == Wildcard && index != len(segments)-1) || !validSegment(segment) {
			return false
		}
	}

	return true
}

// Concrete reports whether a node is valid and contains no wildcard.
func (node Node) Concrete() bool {
	return node.Valid() && !strings.Contains(string(node), Wildcard)
}

// Matches reports whether a granted node covers a concrete queried node.
func (node Node) Matches(query Node) bool {
	if !node.Valid() || !query.Concrete() {
		return false
	}
	if node == Node(Wildcard) || node == query {
		return true
	}

	prefix := strings.TrimSuffix(string(node), "."+Wildcard)
	return prefix != string(node) && strings.HasPrefix(string(query), prefix+".")
}

// Specificity returns the number of fixed segments covering a query or minus one.
func (node Node) Specificity(query Node) int {
	if !node.Matches(query) {
		return -1
	}
	if node == Node(Wildcard) {
		return 0
	}

	return len(strings.Split(strings.TrimSuffix(string(node), "."+Wildcard), "."))
}

// validSegment reports whether a dotted node segment is accepted.
func validSegment(segment string) bool {
	if segment == Wildcard {
		return true
	}
	for _, character := range segment {
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '_' && character != '-' {
			return false
		}
	}

	return segment != ""
}
