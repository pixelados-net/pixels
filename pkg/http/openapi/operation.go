package openapi

import "net/http"

// operation describes one documented route.
type operation struct {
	// method stores the HTTP method.
	method string
	// path stores the OpenAPI path pattern.
	path string
	// tag stores the Scalar section.
	tag string
	// summary stores the operation summary.
	summary string
	// description stores the operation description.
	description string
	// request stores the reflected request model.
	request any
	// responses stores reflected response models.
	responses []response
	// secured reports whether API key auth is required.
	secured bool
}

// response describes one documented response.
type response struct {
	// status stores the HTTP response status.
	status int
	// body stores the reflected response body.
	body any
	// description stores the response description.
	description string
	// contentType stores the optional response content type.
	contentType string
}

// operations returns all documented routes.
func operations() []operation {
	items := publicOperations()
	items = append(items, ssoOperations()...)
	items = append(items, adminOperations()...)
	items = append(items, roomVoteOperations()...)
	items = append(items, fallbackOperation())

	return items
}

// jsonResponse creates an application/json response.
func jsonResponse(status int, body any, description string) response {
	return response{status: status, body: body, description: description, contentType: "application/json"}
}

// emptyResponse creates a response without content.
func emptyResponse(status int, description string) response {
	return response{status: status, description: description}
}

// errorResponses returns common JSON error responses.
func errorResponses(statuses ...int) []response {
	responses := make([]response, 0, len(statuses))
	for _, status := range statuses {
		responses = append(responses, jsonResponse(status, &ErrorResponse{}, http.StatusText(status)))
	}

	return responses
}
