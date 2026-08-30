package app

import (
	"context"
	"youtrack_backend/internal/api/fields"
	"youtrack_backend/internal/model"
)

// GetProject يعيد مشروعًا بمعرّفه أو رمزه المختصر مع جلب الحقول المطلوبة.
func (a *App) GetProject(ctx context.Context, id string, tree *fields.FieldTree) (*model.Project, error) {
	if id == "" {
		return nil, model.BadRequest("project id required")
	}
	p, err := a.store.Projects().GetDetailed(ctx, id, tree)
	if err != nil {
		return nil, model.NotFound("project %q not found", id)
	}
	p.Normalize()
	return p, nil
}

// GetProjectTeamAndLeader يعيد القائد والفريق فقط (الطلب #27) مع تقليم الأعمدة.
func (a *App) GetProjectTeamAndLeader(ctx context.Context, id string, leaderTree, teamTree *fields.FieldTree) (*model.ProjectTeamAndLeader, error) {
	if id == "" {
		return nil, model.BadRequest("project id required")
	}
	res, err := a.store.Projects().GetProjectTeamAndLeader(ctx, id, leaderTree, teamTree)
	if err != nil {
		return nil, model.NotFound("project %q not found", id)
	}
	res.Normalize()
	return res, nil
}

// GetAllProjects يعيد قائمة المشاريع.
func (a *App) GetAllProjects(ctx context.Context) ([]*model.Project, error) {
	projects, err := a.store.Projects().All(ctx)
	if err != nil {
		return nil, model.Internal("failed to load projects: %v", err)
	}
	for _, p := range projects {
		p.Normalize()
	}
	return projects, nil
}
