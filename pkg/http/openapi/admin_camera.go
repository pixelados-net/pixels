package openapi

import (
	"net/http"

	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
	cameraroutes "github.com/niflaot/pixels/pkg/http/camera/routes"
)

// adminCameraOperations returns protected camera administration operations.
func adminCameraOperations() []operation {
	return []operation{
		adminCamera(http.MethodGet, "/api/admin/camera/settings", "Read camera settings", &CameraActorRequest{}, &cameraroutes.SettingsResponse{}, http.StatusOK),
		adminCamera(http.MethodPut, "/api/admin/camera/settings", "Update camera settings", &cameraroutes.SettingsRequest{}, &cameraroutes.SettingsResponse{}, http.StatusOK),
		adminCamera(http.MethodGet, "/api/admin/camera/captures/{playerId}", "List player camera captures", &CameraPlayerRequest{}, &[]camerarecord.Capture{}, http.StatusOK),
		adminCamera(http.MethodGet, "/api/admin/camera/gallery", "List camera gallery", &CameraGalleryRequest{}, &[]camerarecord.Publication{}, http.StatusOK),
		adminCamera(http.MethodDelete, "/api/admin/camera/gallery/{publicationId}", "Remove camera publication", &CameraPublicationDeleteRequest{}, nil, http.StatusNoContent),
		adminCamera(http.MethodDelete, "/api/admin/camera/photos/{itemId}", "Delete photo furniture", &CameraPhotoDeleteRequest{}, nil, http.StatusNoContent),
	}
}

// adminCamera creates one protected camera administration operation.
func adminCamera(method string, path string, summary string, request any, body any, status int) operation {
	responses := errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError)
	if body == nil {
		responses = append([]response{emptyResponse(status, summary+".")}, responses...)
	} else {
		responses = append([]response{jsonResponse(status, body, summary+".")}, responses...)
	}
	return operation{method: method, path: path, tag: "Camera", summary: summary, description: summary + ". Mutations require actorPlayerId and reason for durable audit.", request: request, responses: responses, secured: true}
}

// CameraActorRequest documents read attribution.
type CameraActorRequest struct {
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `header:"X-Actor-Player-ID" required:"true"`
}

// CameraPlayerRequest documents one player capture query.
type CameraPlayerRequest struct {
	CameraActorRequest
	// PlayerID identifies the target player.
	PlayerID int64 `path:"playerId"`
	// Limit bounds returned captures.
	Limit int `query:"limit"`
}

// CameraGalleryRequest documents one gallery query.
type CameraGalleryRequest struct {
	CameraActorRequest
	// Limit bounds returned publications.
	Limit int `query:"limit"`
	// Offset skips publications.
	Offset int `query:"offset"`
	// IncludeRemoved includes moderated rows.
	IncludeRemoved bool `query:"includeRemoved"`
}

// CameraPublicationDeleteRequest documents one publication removal.
type CameraPublicationDeleteRequest struct {
	// PublicationID identifies the gallery row.
	PublicationID int64 `path:"publicationId"`
	cameraroutes.AuditRequest
}

// CameraPhotoDeleteRequest documents one photo furniture deletion.
type CameraPhotoDeleteRequest struct {
	// ItemID identifies the furniture item.
	ItemID int64 `path:"itemId"`
	cameraroutes.AuditRequest
}
