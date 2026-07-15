package openapi

import "github.com/swaggest/openapi-go/openapi3"

const apiKeySecurity = "ApiKeyAuth"

// configureSpec sets top-level document metadata.
func configureSpec(spec *openapi3.Spec) {
	spec.WithOpenapi("3.0.3")
	spec.Info.
		WithTitle("Pixels API").
		WithVersion("0.1.0").
		WithDescription("Public status, websocket entrypoint, development documentation, and protected emulator endpoints.")
	spec.WithServers(openapi3.Server{
		URL: "http://{host}:{port}",
		Variables: map[string]openapi3.ServerVariable{
			"host": {Default: "127.0.0.1"},
			"port": {Default: "3000"},
		},
	})
	spec.WithTags(routeTags()...)
	spec.ComponentsEns().SecuritySchemesEns().WithMapOfSecuritySchemeOrRefValuesItem(
		apiKeySecurity,
		openapi3.SecuritySchemeOrRef{
			SecurityScheme: (&openapi3.SecurityScheme{}).WithAPIKeySecurityScheme(openapi3.APIKeySecurityScheme{
				Name:        "X-API-Key",
				In:          openapi3.APIKeySecuritySchemeInHeader,
				Description: stringPointer("Access key configured by PIXELS_ACCESS_KEY."),
			}),
		},
	)
}

// routeTags returns Scalar route sections.
func routeTags() []openapi3.Tag {
	return []openapi3.Tag{
		tag("Public", "Public health and documentation routes."),
		tag("WebSocket", "Pixel-protocol WebSocket entrypoints."),
		tag("Client Config", "Public Nitro configuration and localized text extensions."),
		tag("SSO", "Protected single-sign-on ticket routes."),
		tag("Admin Players", "Protected player identity and profile administration routes."),
		tag("Admin Bots", "Protected bot inspection, recovery, and bartender configuration routes."),
		tag("Admin Connections", "Protected connection inspection and disconnection routes."),
		tag("Admin Notifications", "Protected localized player communication routes."),
		tag("Admin Currencies", "Protected player currency balance routes."),
		tag("Admin Catalog", "Protected catalog page and offer routes."),
		tag("Admin Subscriptions", "Protected membership, club offer, targeted offer, and calendar routes."),
		tag("Admin Chat", "Protected global filters, bubble thresholds, and bounded history."),
		tag("Admin Permissions", "Protected permission group, membership, and grant routes."),
		tag("Admin Rooms", "Protected room metadata and runtime routes."),
		tag("Admin Room Votes", "Protected permanent room upvote routes."),
		tag("Admin Navigator", "Protected navigator metadata routes."),
		tag("Admin Trading", "Protected Marketplace and direct-trade administration routes."),
		tag("Admin Moderation", "Protected punishments, issues, call-for-help policy, and sanction escalation routes."),
		tag("Fallback", "Authenticated fallback behavior."),
	}
}

// tag creates an OpenAPI tag.
func tag(name string, description string) openapi3.Tag {
	return openapi3.Tag{Name: name, Description: &description}
}

// stringPointer returns a pointer to a string value.
func stringPointer(value string) *string {
	return &value
}
