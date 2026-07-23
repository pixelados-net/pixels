package openapi

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestSpecIsJSON verifies the OpenAPI document is valid JSON.
func TestSpecIsJSON(t *testing.T) {
	var document map[string]any

	if err := json.Unmarshal(Bytes(), &document); err != nil {
		t.Fatalf("unmarshal spec: %v", err)
	}

	if document["openapi"] != "3.0.3" {
		t.Fatalf("expected OpenAPI 3.0.3, got %v", document["openapi"])
	}
}

// TestSpecDocumentsRoutes verifies the expected public routes are documented.
func TestSpecDocumentsRoutes(t *testing.T) {
	var document struct {
		Paths map[string]any `json:"paths"`
	}

	if err := json.Unmarshal(Bytes(), &document); err != nil {
		t.Fatalf("unmarshal spec: %v", err)
	}

	for _, path := range []string{
		"/status", "/ws", "/docs", "/*",
		"/api/admin/notifications/send", "/api/admin/currencies/wallet",
		"/api/admin/players", "/api/admin/players/{id}",
		"/api/admin/players/by-username/{username}",
		"/api/admin/players/{playerId}/effects", "/api/admin/players/{playerId}/effects/{effectId}",
		"/api/admin/bots/serve-items", "/api/admin/bots/serve-items/{id}",
		"/api/admin/bots/{id}", "/api/admin/bots/{id}/force-pickup",
		"/api/admin/pets", "/api/admin/pets/{id}",
		"/api/admin/pets/species", "/api/admin/pets/breeds", "/api/admin/pets/commands",
		"/api/admin/pets/reference/refresh", "/api/admin/pets/{id}/owner",
		"/api/admin/pets/{id}/location", "/api/admin/pets/{id}/stats",
		"/api/admin/currencies/grant", "/api/admin/currencies/deduct", "/api/admin/currencies/set",
		"/api/admin/permissions/nodes", "/api/admin/permissions/groups",
		"/api/admin/permissions/groups/{id}/nodes/{node}",
		"/api/admin/permissions/players/{playerId}/groups/{groupId}",
		"/api/admin/permissions/players/{playerId}/effective",
		"/api/admin/permissions/players/{playerId}/check",
		"/api/admin/players/{playerId}/punishments",
		"/api/admin/punishments/{id}",
		"/api/admin/moderation/issues",
		"/api/admin/moderation/cfh-topics",
		"/api/admin/moderation/presets",
		"/api/admin/moderation/sanction-ladder",
	} {
		if _, ok := document.Paths[path]; !ok {
			t.Fatalf("expected path %s to be documented", path)
		}
	}
}

// TestSpecOmitsPlayerCapabilityRoutes verifies capabilities do not collect below players.
func TestSpecOmitsPlayerCapabilityRoutes(t *testing.T) {
	var document struct {
		Paths map[string]any `json:"paths"`
	}

	if err := json.Unmarshal(Bytes(), &document); err != nil {
		t.Fatalf("unmarshal spec: %v", err)
	}

	for _, path := range []string{
		"/api/admin/players/{id}/notifications",
		"/api/admin/players/{id}/currencies",
		"/api/admin/players/{id}/currencies/{type}/grant",
	} {
		if _, ok := document.Paths[path]; ok {
			t.Fatalf("unexpected player capability path %s", path)
		}
	}
}

// TestSpecGroupsRoutes verifies Scalar route sections are documented.
func TestSpecGroupsRoutes(t *testing.T) {
	var document struct {
		Tags  []map[string]string           `json:"tags"`
		Paths map[string]map[string]opGroup `json:"paths"`
	}

	if err := json.Unmarshal(Bytes(), &document); err != nil {
		t.Fatalf("unmarshal spec: %v", err)
	}

	if !hasTag(document.Tags, "Connections") {
		t.Fatal("expected connection tag")
	}
	if !hasTag(document.Tags, "Permissions") {
		t.Fatal("expected permission tag")
	}
	if !hasTag(document.Tags, "Players") {
		t.Fatal("expected player tag")
	}
	if !hasTag(document.Tags, "Moderation") {
		t.Fatal("expected moderation tag")
	}
	if !hasTag(document.Tags, "Pets") {
		t.Fatal("expected pets tag")
	}

	groups := document.Paths["/api/admin/connections"]["get"].Tags
	if len(groups) != 1 || groups[0] != "Connections" {
		t.Fatalf("expected connection route group, got %#v", groups)
	}
}

// TestSpecUsesFocusedDeclaredTags verifies every operation appears in a concise documented section.
func TestSpecUsesFocusedDeclaredTags(t *testing.T) {
	var document struct {
		Tags  []map[string]string           `json:"tags"`
		Paths map[string]map[string]opGroup `json:"paths"`
	}

	if err := json.Unmarshal(Bytes(), &document); err != nil {
		t.Fatalf("unmarshal spec: %v", err)
	}

	declared := make(map[string]struct{}, len(document.Tags))
	for _, item := range document.Tags {
		name := item["name"]
		if strings.HasPrefix(name, "Admin") {
			t.Fatalf("unexpected admin prefix in declared tag %q", name)
		}
		declared[name] = struct{}{}
	}

	for path, methods := range document.Paths {
		for method, operation := range methods {
			if len(operation.Tags) != 1 {
				t.Fatalf("expected one tag for %s %s, got %#v", method, path, operation.Tags)
			}
			if _, ok := declared[operation.Tags[0]]; !ok {
				t.Fatalf("operation %s %s uses undeclared tag %q", method, path, operation.Tags[0])
			}
			if strings.HasPrefix(operation.Tags[0], "Admin") {
				t.Fatalf("unexpected admin prefix for %s %s", method, path)
			}
		}
	}
}

// TestSpecSeparatesLargeRouteFamilies verifies representative routes stay in focused sections.
func TestSpecSeparatesLargeRouteFamilies(t *testing.T) {
	var document struct {
		Paths map[string]map[string]opGroup `json:"paths"`
	}

	if err := json.Unmarshal(Bytes(), &document); err != nil {
		t.Fatalf("unmarshal spec: %v", err)
	}

	cases := []struct {
		path   string
		method string
		tag    string
	}{
		{path: "/api/admin/bots/{id}", method: "get", tag: "Bots"},
		{path: "/api/admin/pets/{id}", method: "get", tag: "Pets"},
		{path: "/api/admin/pets/species", method: "get", tag: "Pet Reference"},
		{path: "/api/admin/camera/settings", method: "get", tag: "Camera"},
		{path: "/api/admin/camera/gallery", method: "get", tag: "Photo Gallery"},
		{path: "/api/admin/progression/achievements", method: "get", tag: "Achievements"},
		{path: "/api/admin/progression/quests", method: "get", tag: "Quests"},
		{path: "/api/admin/groups/{groupId}/members", method: "get", tag: "Group Members"},
		{path: "/api/admin/groups/{groupId}/forum/threads", method: "get", tag: "Group Forums"},
		{path: "/api/admin/catalog/items", method: "get", tag: "Catalog Offers"},
		{path: "/api/admin/subscriptions/club-offers", method: "get", tag: "Club Offers"},
		{path: "/api/admin/marketplace/listings/{id}/force-close", method: "post", tag: "Marketplace"},
	}

	for _, testCase := range cases {
		operation, ok := document.Paths[testCase.path][testCase.method]
		if !ok {
			t.Fatalf("missing operation %s %s", testCase.method, testCase.path)
		}
		if len(operation.Tags) != 1 || operation.Tags[0] != testCase.tag {
			t.Fatalf("expected %s %s in %q, got %#v", testCase.method, testCase.path, testCase.tag, operation.Tags)
		}
	}
}

// opGroup contains the OpenAPI operation groups needed by tests.
type opGroup struct {
	// Tags stores the operation tag list.
	Tags []string `json:"tags"`
}

// hasTag reports whether the OpenAPI document contains a tag.
func hasTag(tags []map[string]string, name string) bool {
	for _, tag := range tags {
		if tag["name"] == name {
			return true
		}
	}

	return false
}
