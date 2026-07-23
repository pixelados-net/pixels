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

// tag creates an OpenAPI tag.
func tag(name string, description string) openapi3.Tag {
	return openapi3.Tag{Name: name, Description: &description}
}

// stringPointer returns a pointer to a string value.
func stringPointer(value string) *string {
	return &value
}
