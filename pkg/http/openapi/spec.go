// Package openapi contains the documented Pixels HTTP route contract.
package openapi

import (
	"encoding/json"

	oapi "github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
)

// Spec is the OpenAPI document for the HTTP surface.
var Spec = mustBuild()

// Bytes returns the OpenAPI document bytes.
func Bytes() []byte {
	return []byte(Spec)
}

// build creates the OpenAPI document.
func build() (string, error) {
	reflector := openapi3.NewReflector()
	configureSpec(reflector.SpecEns())
	for _, operation := range operations() {
		if err := addOperation(reflector, operation); err != nil {
			return "", err
		}
	}

	data, err := json.MarshalIndent(reflector.Spec, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// mustBuild creates the OpenAPI document or panics during startup.
func mustBuild() string {
	spec, err := build()
	if err != nil {
		panic(err)
	}

	return spec
}

// addOperation reflects one documented operation.
func addOperation(reflector *openapi3.Reflector, item operation) error {
	context, err := reflector.NewOperationContext(item.method, item.path)
	if err != nil {
		return err
	}

	context.SetTags(item.tag)
	context.SetSummary(item.summary)
	context.SetDescription(item.description)
	if item.secured {
		context.AddSecurity(apiKeySecurity)
	}
	if item.request != nil {
		context.AddReqStructure(item.request)
	}
	for _, response := range item.responses {
		context.AddRespStructure(response.body, responseOptions(response)...)
	}

	return reflector.AddOperation(context)
}

// responseOptions returns OpenAPI content options for a response.
func responseOptions(response response) []oapi.ContentOption {
	return []oapi.ContentOption{
		oapi.WithHTTPStatus(response.status),
		func(unit *oapi.ContentUnit) {
			unit.ContentType = response.contentType
			unit.Description = response.description
		},
	}
}
