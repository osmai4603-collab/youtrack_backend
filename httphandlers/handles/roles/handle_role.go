package roleshandler

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"youtrack_backend/internal/httphandlers/response"
	platErrors "youtrack_backend/internal/platform/errors"
	permstorage "youtrack_backend/internal/storage/permissions"

	"github.com/go-chi/chi/v5"
)

type updateRoleRequest struct {
	Name        json.RawMessage `json:"name"`
	Description json.RawMessage `json:"description"`
}

func (handler *Handler) handleCreate(writer http.ResponseWriter, request *http.Request) {
	var payload updateRoleRequest
	request.Body = http.MaxBytesReader(writer, request.Body, 1<<20)
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		response.Error(writer, request, platErrors.Invalid("roles.Create", "invalid request body", err))
		return
	}
	name, err := decodeOptionalString(payload.Name)
	if err != nil {
		response.Error(writer, request, platErrors.Invalid("roles.Create", "invalid role name", err))
		return
	}
	description, err := decodeOptionalString(payload.Description)
	if err != nil {
		response.Error(writer, request, platErrors.Invalid("roles.Create", "invalid role description", err))
		return
	}
	if name == nil {
		response.Error(writer, request, platErrors.Invalid("roles.Create", "role name is required", errors.New("missing role name")))
		return
	}
	role, err := handler.roles.Create(request.Context(), *name, description)
	if err != nil {
		handler.writeError(writer, request, "roles.Create", err)
		return
	}
	response.Created(writer, role)
}

func (handler *Handler) handleGet(writer http.ResponseWriter, request *http.Request) {
	roleID := strings.TrimSpace(chi.URLParam(request, "roleID"))
	role, err := handler.roles.Get(request.Context(), roleID)
	if err != nil {
		handler.writeError(writer, request, "roles.Get", err)
		return
	}
	response.JSON(writer, http.StatusOK, role)
}

func (handler *Handler) handleList(writer http.ResponseWriter, request *http.Request) {
	roles, err := handler.roles.List(request.Context())
	if err != nil {
		handler.writeError(writer, request, "roles.List", err)
		return
	}
	response.JSON(writer, http.StatusOK, roles)
}

func (handler *Handler) handleUpdate(writer http.ResponseWriter, request *http.Request) {
	var payload updateRoleRequest
	request.Body = http.MaxBytesReader(writer, request.Body, 1<<20)
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		response.Error(writer, request, platErrors.Invalid("roles.Update", "invalid request body", err))
		return
	}
	name, err := decodeOptionalString(payload.Name)
	if err != nil {
		response.Error(writer, request, platErrors.Invalid("roles.Update", "invalid role name", err))
		return
	}
	description, err := decodeOptionalString(payload.Description)
	if err != nil {
		response.Error(writer, request, platErrors.Invalid("roles.Update", "invalid role description", err))
		return
	}

	roleID := strings.TrimSpace(chi.URLParam(request, "roleID"))
	role, err := handler.roles.Update(request.Context(), roleID, name, description)
	if err != nil {
		handler.writeError(writer, request, "roles.Update", err)
		return
	}
	response.JSON(writer, http.StatusOK, role)
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
	roleID := strings.TrimSpace(chi.URLParam(request, "roleID"))
	if err := handler.roles.Delete(request.Context(), roleID); err != nil {
		handler.writeError(writer, request, "roles.Delete", err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) writeError(writer http.ResponseWriter, request *http.Request, op string, err error) {
	if errors.Is(err, permstorage.ErrImmutableRole) {
		err = platErrors.Forbidden(op, "immutable roles cannot be changed", err)
	}
	handler.logger.LogAttrs(request.Context(), slog.LevelWarn, "role operation failed",
		slog.String("operation", op),
		slog.String("role_id", chi.URLParam(request, "roleID")),
		slog.Any("error", err),
	)
	response.Error(writer, request, err)
}
