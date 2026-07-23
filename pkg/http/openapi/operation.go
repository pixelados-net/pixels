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
	items = append(items, adminPlayerOperations()...)
	items = append(items, adminBotOperations()...)
	items = append(items, adminPetOperations()...)
	items = append(items, adminCraftingOperations()...)
	items = append(items, adminCameraOperations()...)
	items = append(items, adminProgressionOperations()...)
	items = append(items, adminGamesOperations()...)
	items = append(items, adminGroupOperations()...)
	items = append(items, adminOperations()...)
	items = append(items, wiredAdminOperations()...)
	items = append(items, roomVoteOperations()...)
	items = append(items, adminTradingOperations()...)
	items = append(items, moderationAdminOperations()...)
	items = append(items, fallbackOperation())
	for index := range items {
		items[index].tag = operationCategory(items[index].path, items[index].tag)
	}

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

// adminSubscription creates one protected subscription administration operation.
func adminSubscription(method string, path string, summary string, request any, body any, status int) operation {
	responses := errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError)
	if body == nil {
		responses = append([]response{emptyResponse(status, summary+".")}, responses...)
	} else {
		responses = append([]response{jsonResponse(status, body, summary+".")}, responses...)
	}

	return operation{method: method, path: path, tag: "Subscriptions", summary: summary,
		description: summary + ".", request: request, responses: responses, secured: true}
}

// adminMessenger creates a protected messenger administration operation.
func adminMessenger(method string, path string, summary string, request any, body any, status int) operation {
	return operation{method: method, path: path, tag: "Friends & Privacy", summary: summary,
		description: summary + ".", request: request,
		responses: append([]response{jsonResponse(status, body, summary+".")},
			errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError)...), secured: true}
}

// adminChat creates a protected chat administration operation.
func adminChat(method string, path string, summary string, request any, body any, status int) operation {
	responses := errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusInternalServerError)
	if body == nil {
		responses = append([]response{emptyResponse(status, summary+".")}, responses...)
	} else {
		responses = append([]response{jsonResponse(status, body, summary+".")}, responses...)
	}

	return operation{method: method, path: path, tag: "Chat", summary: summary,
		description: summary + ".", request: request, responses: responses, secured: true}
}

// adminCatalog creates a catalog administration operation.
func adminCatalog(method string, path string, summary string, request any, body any) operation {
	responses := errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError)
	if body == nil {
		responses = append([]response{emptyResponse(http.StatusNoContent, summary+".")}, responses...)
	} else {
		responses = append([]response{jsonResponse(http.StatusOK, body, summary+".")}, responses...)
	}

	return operation{method: method, path: path, tag: "Catalog", summary: summary,
		description: summary + ".", request: request, responses: responses, secured: true}
}
