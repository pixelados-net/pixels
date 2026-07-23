package openapi

import "net/http"

// adminRoomRead creates a read-only room admin operation.
func adminRoomRead(path string, summary string, request any, body any) operation {
	return adminTaggedRead("Rooms", path, summary, request, body)
}

// adminRoomMutation creates one protected room administration mutation.
func adminRoomMutation(method string, path string, summary string, request any, body any) operation {
	responses := errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError)
	if body == nil {
		responses = append([]response{emptyResponse(http.StatusNoContent, summary+".")}, responses...)
	} else {
		responses = append([]response{jsonResponse(http.StatusOK, body, summary+".")}, responses...)
	}
	return operation{method: method, path: path, tag: "Rooms", summary: summary, description: summary + ".", request: request, responses: responses, secured: true}
}

// adminNavigatorRead creates a read-only navigator admin operation.
func adminNavigatorRead(path string, summary string, request any, body any) operation {
	return adminTaggedRead("Navigator", path, summary, request, body)
}

// adminNavigatorMutation creates one protected navigator administration mutation.
func adminNavigatorMutation(method string, path string, summary string, request any, body any) operation {
	responses := errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError)
	if body == nil {
		responses = append([]response{emptyResponse(http.StatusNoContent, summary+".")}, responses...)
	} else {
		responses = append([]response{jsonResponse(http.StatusOK, body, summary+".")}, responses...)
	}
	return operation{method: method, path: path, tag: "Navigator", summary: summary, description: summary + ".", request: request, responses: responses, secured: true}
}

// adminTaggedRead creates a tagged read-only admin operation.
func adminTaggedRead(tag string, path string, summary string, request any, body any) operation {
	return operation{
		method: http.MethodGet, path: path, tag: tag, summary: summary, description: summary + ".", request: request,
		responses: append([]response{jsonResponse(http.StatusOK, body, summary+".")}, errorResponses(http.StatusUnauthorized, http.StatusNotFound)...),
		secured:   true,
	}
}
