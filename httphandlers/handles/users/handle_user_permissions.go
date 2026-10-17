package usershandler

import (
	"net/http"

	"youtrack_backend/internal/httphandlers/middlewares"
	"youtrack_backend/internal/httphandlers/response"
	platErrors "youtrack_backend/internal/platform/errors"
)

type RoleAssignmentParams struct {
	RoleID         string   `json:"roleId"`
	RoleName       string   `json:"roleName"`
	ScopeType      string   `json:"scopeType"`
	ProjectID      string   `json:"projectId"`
	OrganizationID string   `json:"organizationId"`
	GroupID        string   `json:"groupId"`
	Permissions    []string `json:"permissions"`
	Type           string   `json:"$type"`
}

type PermissionsParams struct {
	RoleAssignments []RoleAssignmentParams `json:"roleAssignments"`
	OwnedProjectIDs []string               `json:"ownedProjectIds"`
}

func (handler *Handler) handleUserPermissions(writer http.ResponseWriter, request *http.Request) {
	userID, ok := middlewares.GetCurrentUserID(request.Context())
	if !ok || userID == "" {
		response.Error(writer, platErrors.Unauthorized("userPermissions", "authentication required", nil))
		return
	}

	result, err := handler.permissions.Get(request.Context(), userID)
	if err != nil {
		response.Error(writer, platErrors.Internal("userPermissions", err.Error(), err))
		return
	}

	assignments := make([]RoleAssignmentParams, 0, len(result.RoleAssignments))
	for _, assignment := range result.RoleAssignments {
		assignments = append(assignments, RoleAssignmentParams{
			RoleID:         assignment.RoleID,
			RoleName:       assignment.RoleName,
			ScopeType:      assignment.ScopeType,
			ProjectID:      assignment.ProjectID,
			OrganizationID: assignment.OrganizationID,
			GroupID:        assignment.GroupID,
			Permissions:    assignment.Permissions,
			Type:           "UserRoleAssignment",
		})
	}

	response.JSON(writer, http.StatusOK, PermissionsParams{
		RoleAssignments: assignments,
		OwnedProjectIDs: result.OwnedProjectIDs,
	})
}
