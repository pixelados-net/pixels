package textfilter

import "unicode"

// Matcher is an immutable Aho-Corasick text matcher.
type Matcher struct {
	// nodes stores compiled automaton states.
	nodes []node
	// maxPattern stores the largest normalized pattern length.
	maxPattern int
}

// node stores one Aho-Corasick automaton state.
type node struct {
	// transitions maps normalized runes to successor states.
	transitions map[rune]int
	// failure stores the longest available suffix state.
	failure int
	// outputs stores normalized pattern lengths ending at this state.
	outputs []int
}

// Compile builds an immutable matcher from normalized filter patterns.
func Compile(words []string) *Matcher {
	matcher := &Matcher{nodes: []node{{transitions: make(map[rune]int)}}}
	seen := make(map[string]struct{}, len(words))
	for _, word := range words {
		pattern := normalize(word)
		if len(pattern) == 0 {
			continue
		}
		key := string(pattern)
		if _, found := seen[key]; found {
			continue
		}
		seen[key] = struct{}{}
		matcher.insert(pattern)
	}
	matcher.failures()

	return matcher
}

// insert adds one normalized pattern to the trie.
func (matcher *Matcher) insert(pattern []rune) {
	state := 0
	for _, value := range pattern {
		next, found := matcher.nodes[state].transitions[value]
		if !found {
			next = len(matcher.nodes)
			matcher.nodes = append(matcher.nodes, node{transitions: make(map[rune]int)})
			matcher.nodes[state].transitions[value] = next
		}
		state = next
	}
	matcher.nodes[state].outputs = append(matcher.nodes[state].outputs, len(pattern))
	if len(pattern) > matcher.maxPattern {
		matcher.maxPattern = len(pattern)
	}
}

// failures builds suffix links with breadth-first traversal.
func (matcher *Matcher) failures() {
	queue := make([]int, 0, len(matcher.nodes))
	for _, child := range matcher.nodes[0].transitions {
		queue = append(queue, child)
	}
	for offset := 0; offset < len(queue); offset++ {
		state := queue[offset]
		for value, child := range matcher.nodes[state].transitions {
			queue = append(queue, child)
			failure := matcher.nodes[state].failure
			for failure != 0 {
				if _, found := matcher.nodes[failure].transitions[value]; found {
					break
				}
				failure = matcher.nodes[failure].failure
			}
			if target, found := matcher.nodes[failure].transitions[value]; found && target != child {
				matcher.nodes[child].failure = target
			}
			inherited := matcher.nodes[matcher.nodes[child].failure].outputs
			matcher.nodes[child].outputs = append(matcher.nodes[child].outputs, inherited...)
		}
	}
}

// next advances one normalized rune through failure transitions.
func (matcher *Matcher) next(state int, value rune) int {
	for state != 0 {
		if next, found := matcher.nodes[state].transitions[value]; found {
			return next
		}
		state = matcher.nodes[state].failure
	}
	if next, found := matcher.nodes[0].transitions[value]; found {
		return next
	}

	return 0
}

// normalize removes separators and folds one pattern to lowercase runes.
func normalize(value string) []rune {
	result := make([]rune, 0, len(value))
	for _, current := range value {
		if unicode.IsLetter(current) || unicode.IsNumber(current) {
			result = append(result, unicode.ToLower(current))
		}
	}

	return result
}
