package permission

import (
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

// Registration describes one domain-owned permission node.
type Registration struct {
	// Node stores the registered permission identifier.
	Node Node
	// PerkName stores an optional Nitro-visible perk code.
	PerkName string
	// Package stores the declaring package path relative to the repository.
	Package string
}

var (
	// registryMutex protects process-wide node registration.
	registryMutex sync.RWMutex
	// registrations stores process-wide node metadata by identifier.
	registrations = make(map[Node]Registration)
)

// RegisterNode declares one permission node and its optional Nitro perk.
func RegisterNode(node Node, perkName string) Node {
	if !node.Valid() {
		panic("invalid permission node: " + string(node))
	}

	registration := Registration{Node: node, PerkName: strings.TrimSpace(perkName), Package: callerPackage()}
	registryMutex.Lock()
	defer registryMutex.Unlock()
	if _, exists := registrations[node]; exists {
		panic("duplicate permission node: " + string(node))
	}
	registrations[node] = registration

	return node
}

// AllNodes returns registered node identifiers in stable lexical order.
func AllNodes() []Node {
	metadata := RegisteredNodes()
	nodes := make([]Node, 0, len(metadata))
	for _, registration := range metadata {
		nodes = append(nodes, registration.Node)
	}

	return nodes
}

// RegisteredNodes returns registered node metadata in stable lexical order.
func RegisteredNodes() []Registration {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	metadata := make([]Registration, 0, len(registrations))
	for _, registration := range registrations {
		metadata = append(metadata, registration)
	}
	sort.Slice(metadata, func(left int, right int) bool {
		return metadata[left].Node < metadata[right].Node
	})

	return metadata
}

// Registered reports whether a node is declared by a domain package.
func Registered(node Node) bool {
	registryMutex.RLock()
	defer registryMutex.RUnlock()
	_, found := registrations[node]

	return found
}

// callerPackage resolves the package declaring a registered node.
func callerPackage() string {
	_, file, _, found := runtime.Caller(2)
	if !found {
		return "unknown"
	}
	clean := filepath.ToSlash(file)
	if index := strings.LastIndex(clean, "/internal/"); index >= 0 {
		return strings.TrimSuffix(clean[index+1:], "/"+filepath.Base(clean))
	}

	return filepath.ToSlash(filepath.Dir(file))
}
