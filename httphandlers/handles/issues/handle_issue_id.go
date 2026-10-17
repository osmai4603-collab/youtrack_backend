package issueshandler

import (
	"errors"
	"net/http"
	"youtrack_backend/internal/httphandlers/middlewares"
	"youtrack_backend/internal/httphandlers/response"
	platErrors "youtrack_backend/internal/platform/errors"
	issueusecase "youtrack_backend/internal/usecases/issue"

	"github.com/go-chi/chi/v5"
)

func (hlr *Handler) handleIssueByID(writer http.ResponseWriter, request *http.Request) {
	userID, ok := middlewares.GetCurrentUserID(request.Context())
	if !ok || userID == "" {
		response.Error(writer, platErrors.Unauthorized("issues", "authentication required",
			errors.New("missing user identity")))
		return
	}

	issueID := chi.URLParam(request, "issueID")
	if issueID == "" {
		writeIssuesError(writer, issueusecase.ErrIssueIDRequired, "issue id is required")
		return
	}

	result, err := hlr.issueByID.Execute(request.Context(), &issueusecase.GetIssueParams{
		UserID:  userID,
		IssueID: issueID,
	})
	if err != nil {
		hlr.logRejected(request, userID, "issue read rejected", err)
		writeIssuesError(writer, err, "failed to read issue")
		return
	}

	response.JSON(writer, http.StatusOK, newIssueParams(result.Issue))
}
