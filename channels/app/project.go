package app

import (
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"
	"youtrack_backend/channels/request"
)

// GetProject يعيد مشروعًا بمعرّفه أو رمزه المختصر مع جلب الحقول المطلوبة.
func (a *YouTrackApp) GetProject(c request.CTX, id string, tree *fields.FieldTree) (*model.Project, *model.AppError) {
	if id == "" {
		return nil, model.NewBadRequestError("App.GetProject", "project id required")
	}
	ctx := c.Context()
	p, err := a.Store().Projects().GetDetailed(ctx, id, tree)
	if err != nil {
		return nil, model.NewNotFoundError("App.GetProject", "project not found")
	}
	p.Normalize()
	return p, nil
}

// GetProjectTeamAndLeader يعيد القائد والفريق فقط (الطلب #27) مع تقليم الأعمدة.
func (a *YouTrackApp) GetProjectTeamAndLeader(c request.CTX, id string, leaderTree, teamTree *fields.FieldTree) (*model.ProjectTeamAndLeader, *model.AppError) {
	if id == "" {
		return nil, model.NewBadRequestError("App.GetProjectTeamAndLeader", "project id required")
	}
	ctx := c.Context()
	res, err := a.Store().Projects().GetProjectTeamAndLeader(ctx, id, leaderTree, teamTree)
	if err != nil {
		return nil, model.NewNotFoundError("App.GetProjectTeamAndLeader", "project not found")
	}
	res.Normalize()
	return res, nil
}

// GetAllProjects يعيد قائمة المشاريع.
func (a *YouTrackApp) GetAllProjects(c request.CTX) ([]*model.Project, *model.AppError) {
	ctx := c.Context()
	projects, err := a.Store().Projects().All(ctx)
	if err != nil {
		return nil, model.NewInternalError("App.GetAllProjects", "failed to load projects", err)
	}
	for _, p := range projects {
		p.Normalize()
	}
	return projects, nil
}
