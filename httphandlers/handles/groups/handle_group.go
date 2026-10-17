package groupshandler

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"youtrack_backend/internal/httphandlers/response"
	platErrors "youtrack_backend/internal/platform/errors"
	groupstorage "youtrack_backend/internal/storage/groups"

	"github.com/go-chi/chi/v5"
)

type updateGroupRequest struct {
	Name        json.RawMessage `json:"name"`
	Description json.RawMessage `json:"description"`
}

type createGroupRequest struct {
	Name          json.RawMessage `json:"name"`
	Description   json.RawMessage `json:"description"`
	GroupType     string          `json:"groupType"`
	ParentGroupID *string         `json:"parentGroupId"`
	AutoJoin      bool            `json:"autoJoin"`
}

func (handler *Handler) handleCreate(writer http.ResponseWriter, request *http.Request) {
	var payload createGroupRequest
	request.Body = http.MaxBytesReader(writer, request.Body, 1<<20)
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		response.Error(writer, request, platErrors.Invalid("groups.Create", "invalid request body", err))
		return
	}
	name, err := decodeOptionalString(payload.Name)
	if err != nil {
		response.Error(writer, request, platErrors.Invalid("groups.Create", "invalid group name", err))
		return
	}
	description, err := decodeOptionalString(payload.Description)
	if err != nil {
		response.Error(writer, request, platErrors.Invalid("groups.Create", "invalid group description", err))
		return
	}
	group, err := handler.groups.Create(request.Context(), name, description, payload.GroupType, payload.ParentGroupID, payload.AutoJoin)
	if err != nil {
		handler.writeError(writer, request, "groups.Create", err)
		return
	}
	response.Created(writer, group)
}

func (handler *Handler) handleGet(writer http.ResponseWriter, request *http.Request) {
	groupID := strings.TrimSpace(chi.URLParam(request, "groupID"))
	group, err := handler.groups.Get(request.Context(), groupID)
	if err != nil {
		handler.writeError(writer, request, "groups.Get", err)
		return
	}
	response.JSON(writer, http.StatusOK, group)
}

func (handler *Handler) handleList(writer http.ResponseWriter, request *http.Request) {
	groups, err := handler.groups.List(request.Context())
	if err != nil {
		handler.writeError(writer, request, "groups.List", err)
		return
	}
	response.JSON(writer, http.StatusOK, groups)
}

func (handler *Handler) handleUpdate(writer http.ResponseWriter, request *http.Request) {
	var payload updateGroupRequest
	request.Body = http.MaxBytesReader(writer, request.Body, 1<<20)
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		response.Error(writer, request, platErrors.Invalid("groups.Update", "invalid request body", err))
		return
	}
	name, err := decodeOptionalString(payload.Name)
	if err != nil {
		response.Error(writer, request, platErrors.Invalid("groups.Update", "invalid group name", err))
		return
	}
	description, err := decodeOptionalString(payload.Description)
	if err != nil {
		response.Error(writer, request, platErrors.Invalid("groups.Update", "invalid group description", err))
		return
	}

	groupID := strings.TrimSpace(chi.URLParam(request, "groupID"))
	group, err := handler.groups.Update(request.Context(), groupID, name, description)
	if err != nil {
		handler.writeError(writer, request, "groups.Update", err)
		return
	}
	response.JSON(writer, http.StatusOK, group)
}

func decodeOptionalString(raw json.RawMessage) (*string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	value := ""
	if !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, err
		}
	}
	return &value, nil
}

func (handler *Handler) handleDelete(writer http.ResponseWriter, request *http.Request) {
	groupID := strings.TrimSpace(chi.URLParam(request, "groupID"))
	successorID := strings.TrimSpace(request.URL.Query().Get("successor"))
	if err := handler.groups.Delete(request.Context(), groupID, successorID); err != nil {
		handler.writeError(writer, request, "groups.Delete", err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) writeError(writer http.ResponseWriter, request *http.Request, op string, err error) {
	switch {
	case errors.Is(err, groupstorage.ErrSystemGroup):
		err = platErrors.Forbidden(op, "system groups cannot be changed or deleted", err)
	case errors.Is(err, groupstorage.ErrNameRequired):
		err = platErrors.Invalid(op, "group name is required", err)
	case errors.Is(err, groupstorage.ErrNotFound):
		err = platErrors.NotFound(op, "group not found", err)
	case errors.Is(err, groupstorage.ErrSuccessorNeeded), errors.Is(err, groupstorage.ErrIDRequired), errors.Is(err, groupstorage.ErrNameRequired):
		err = platErrors.Invalid(op, "invalid group request", err)
	}
	handler.logger.LogAttrs(request.Context(), slog.LevelWarn, "group operation failed",
		slog.String("operation", op),
		slog.String("group_id", chi.URLParam(request, "groupID")),
		slog.Any("error", err),
	)
	response.Error(writer, request, err)
}
