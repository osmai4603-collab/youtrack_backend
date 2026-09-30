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
	if err := a.requireProjectAccess(c, id, "App.GetProject"); err != nil {
		return nil, err
	}
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
	if err := a.requireProjectAccess(c, id, "App.GetProjectTeamAndLeader"); err != nil {
		return nil, err
	}
	res, err := a.Store().Projects().GetProjectTeamAndLeader(ctx, id, leaderTree, teamTree)
	if err != nil {
		return nil, model.NewNotFoundError("App.GetProjectTeamAndLeader", "project not found")
	}
	res.Normalize()
	return res, nil
}

// GetAllProjects يعيد قائمة مشاريع المستخدم الحالي فقط.
func (a *YouTrackApp) GetAllProjects(c request.CTX) ([]*model.Project, *model.AppError) {
	ctx := c.Context()

	// الجلسة الإدارية ترى كل المشاريع، وأي جلسة أخرى ترى مشاريعها فقط.
	var projects []*model.Project
	var err error
	if a.HasGlobalProjectAccess(c) {
		projects, err = a.Store().Projects().All(ctx)
	} else {
		projects, err = a.Store().Projects().AllByUser(ctx, c.UserID())
	}
	if err != nil {
		return nil, model.NewInternalError("App.GetAllProjects", "failed to load projects", err)
	}
	for _, p := range projects {
		p.Normalize()
	}
	return projects, nil
}

// requireProjectAccess يتحقق من أن المشروع ضمن نطاق وصول المستخدم الحالي،
// ويعيد 404 (وليس 403) حتى لا يكشف وجود كيان لا يملكه المستخدم.
func (a *YouTrackApp) requireProjectAccess(c request.CTX, projectRef, where string) *model.AppError {
	if a.HasGlobalProjectAccess(c) {
		return nil
	}
	allowed, err := a.Store().Projects().CanAccessProject(c.Context(), c.UserID(), projectRef)
	if err != nil {
		return model.NewInternalError(where, "failed to check project access", err)
	}
	if !allowed {
		return model.NewNotFoundError(where, "project not found")
	}
	return nil
}
