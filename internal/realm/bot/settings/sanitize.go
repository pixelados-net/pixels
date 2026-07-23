package settings

import (
	stdhtml "html"
	"strings"

	xhtml "golang.org/x/net/html"
)

// sanitizeFixedPoint decodes and strips HTML until stable with a defensive cap.
func sanitizeFixedPoint(value string) (string, bool) {
	for range 5 {
		next := stdhtml.UnescapeString(stripHTML(value))
		if next == value {
			return strings.TrimSpace(next), true
		}
		value = next
	}
	return "", false
}

// stripHTML returns text nodes from one fragment.
func stripHTML(value string) string {
	root, err := xhtml.Parse(strings.NewReader(value))
	if err != nil {
		return ""
	}
	var builder strings.Builder
	var visit func(*xhtml.Node)
	visit = func(node *xhtml.Node) {
		if node.Type == xhtml.TextNode {
			builder.WriteString(node.Data)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(root)
	return builder.String()
}

// truncateRunes returns at most limit Unicode code points.
func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
