package openapi

import (
	"net/http"

	petroutes "github.com/niflaot/pixels/pkg/http/pet/routes"
)

// adminPetOperations returns protected pet administration operations.
func adminPetOperations() []operation {
	return []operation{
		adminPet(http.MethodGet, "/api/admin/pets", "List pets", &PetListRequest{}, &PetListResponse{}, http.StatusOK),
		adminPet(http.MethodPost, "/api/admin/pets", "Create pet", &PetCreateRequest{}, &PetMutationResponse{}, http.StatusCreated),
		adminPet(http.MethodGet, "/api/admin/pets/metrics", "Read pet runtime metrics", &PetActorRequest{}, &petroutes.MetricsResponse{}, http.StatusOK),
		adminPet(http.MethodGet, "/api/admin/pets/species", "List pet species", &PetActorRequest{}, &[]petroutes.SpeciesResponse{}, http.StatusOK),
		adminPet(http.MethodGet, "/api/admin/pets/breeds", "List pet breeds", &PetBreedListRequest{}, &[]petroutes.BreedResponse{}, http.StatusOK),
		adminPet(http.MethodGet, "/api/admin/pets/commands", "List pet commands", &PetActorRequest{}, &petroutes.CommandsResponse{}, http.StatusOK),
		adminPet(http.MethodPost, "/api/admin/pets/reference/refresh", "Refresh pet references", &PetAuditRequest{}, &PetReferenceRefreshResponse{}, http.StatusOK),
		adminPet(http.MethodGet, "/api/admin/pets/{id}", "Read pet", &PetReadRequest{}, &PetAdminResponse{}, http.StatusOK),
		adminPet(http.MethodPatch, "/api/admin/pets/{id}", "Update pet", &PetUpdateRequest{}, &PetMutationResponse{}, http.StatusOK),
		adminPet(http.MethodDelete, "/api/admin/pets/{id}", "Soft delete pet", &PetDeleteRequest{}, nil, http.StatusNoContent),
		adminPet(http.MethodPost, "/api/admin/pets/{id}/owner", "Transfer pet owner", &PetOwnerRequest{}, &PetMutationResponse{}, http.StatusOK),
		adminPet(http.MethodPost, "/api/admin/pets/{id}/location", "Change pet location", &PetLocationRequest{}, &PetMutationResponse{}, http.StatusOK),
		adminPet(http.MethodPost, "/api/admin/pets/{id}/stats", "Change pet stats", &PetStatsRequest{}, &PetMutationResponse{}, http.StatusOK),
	}
}

// adminPet creates one protected pet operation.
func adminPet(method string, path string, summary string, request any, body any, status int) operation {
	responses := errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError)
	if body == nil {
		responses = append([]response{emptyResponse(status, summary+".")}, responses...)
	} else {
		responses = append([]response{jsonResponse(status, body, summary+".")}, responses...)
	}
	return operation{method: method, path: path, tag: "Pets", summary: summary, description: summary + ". Reads resolve X-Actor-Player-ID; mutations resolve actorPlayerId and persist reason for durable audit.", request: request, responses: responses, secured: true}
}
