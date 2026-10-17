package issueshandler

import (
	"errors"
	"net/http"
	"youtrack_backend/internal/httphandlers/middlewares"
	"youtrack_backend/internal/httphandlers/response"
	platErrors "youtrack_backend/internal/platform/errors"
	issueusecase "youtrack_backend/internal/usecases/issue"
)

func (hlr *Handler) handleIssues(writer http.ResponseWriter, request *http.Request) {
	userID, ok := middlewares.GetCurrentUserID(request.Context())
	if !ok || userID == "" {
		response.Error(writer, platErrors.Unauthorized("issues", "authentication required",
			errors.New("missing user identity")))
		return
	}

	query := IssuesQuery(request)

	result, err := hlr.projectIssues.Execute(request.Context(), &issueusecase.ListProjectIssuesParams{
		UserID: userID,
		Query:  query,
	})
	if err != nil {
		hlr.logRejected(request, userID, "issues listing rejected", err)
		writeIssuesError(writer, err, "failed to list issues")
		return
	}

	items := make([]ProjectIssuesResponse, 0, len(result.Issues))
	for _, issue := range result.Issues {
		items = append(items, newIssueParams(issue))
	}

	response.JSON(writer, http.StatusOK, items)
}
