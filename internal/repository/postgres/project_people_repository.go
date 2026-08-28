package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"youtrack_backend/internal/domain"
)

type ProjectPeopleRepository struct {
	db *sql.DB
}

func NewProjectPeopleRepository(db *sql.DB) *ProjectPeopleRepository {
	return &ProjectPeopleRepository{db: db}
}

func (r *ProjectPeopleRepository) GetProjectPeople(projectID string, transitiveRolesQuery string) (*domain.ProjectPeople, error) {
	teamQuery := `SELECT id FROM project_teams WHERE project_id = $1`
	teamRows, err := r.db.Query(teamQuery, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query project teams: %w", err)
	}
	defer teamRows.Close()

	var teamIDs []string
	for teamRows.Next() {
		var teamID string
		if err := teamRows.Scan(&teamID); err != nil {
			return nil, fmt.Errorf("failed to scan team id: %w", err)
		}
		teamIDs = append(teamIDs, teamID)
	}
	if err := teamRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team rows: %w", err)
	}

	people := &domain.ProjectPeople{
		Type: "ProjectPeople",
	}

	if len(teamIDs) == 0 {
		return people, nil
	}

	membersQuery, memberArgs := buildInClause("SELECT user_id FROM project_team_members WHERE team_id IN", teamIDs)
	memberRows, err := r.db.Query(membersQuery, memberArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query team members: %w", err)
	}
	defer memberRows.Close()

	var memberUserIDs []string
	for memberRows.Next() {
		var userID string
		if err := memberRows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan team member user id: %w", err)
		}
		memberUserIDs = append(memberUserIDs, userID)
	}
	if err := memberRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team member rows: %w", err)
	}

	people.TotalUsersInTeamCount = len(memberUserIDs)

	if len(memberUserIDs) == 0 {
		return people, nil
	}

	usersQuery, userArgs := buildInClause(
		`SELECT id, login, email, name, full_name, avatar_url, online, banned, is_locked, is_email_verified, guest
		 FROM users WHERE id IN`, memberUserIDs,
	)
	userRows, err := r.db.Query(usersQuery, userArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query team users: %w", err)
	}
	defer userRows.Close()

	teamUsers := make(map[string]*domain.User)
	for userRows.Next() {
		user := &domain.User{}
		var email, fullName, avatarURL sql.NullString
		if err := userRows.Scan(
			&user.ID,
			&user.Login,
			&email,
			&user.Name,
			&fullName,
			&avatarURL,
			&user.Online,
			&user.Banned,
			&user.IsLocked,
			&user.IsEmailVerified,
			&user.Guest,
		); err != nil {
			return nil, fmt.Errorf("failed to scan team user: %w", err)
		}
		if email.Valid {
			user.Email = email.String
		}
		if fullName.Valid {
			user.FullName = fullName.String
			user.LocalizedName = fullName.String
		}
		if avatarURL.Valid {
			user.AvatarURL = avatarURL.String
		}
		user.Type = "User"
		teamUsers[user.ID] = user
	}
	if err := userRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team user rows: %w", err)
	}

	people.UsersInTeam = make([]*domain.User, 0, len(teamUsers))
	for _, u := range teamUsers {
		people.UsersInTeam = append(people.UsersInTeam, u)
	}

	groupQuery, groupArgs := buildInClause(
		`SELECT DISTINCT ug.id, ug.name, ug.description, ug.all_users_group, ug.icon, ug.is_updatable, ug.is_removable
		 FROM user_groups ug
		 INNER JOIN user_group_members ugm ON ug.id = ugm.group_id
		 WHERE ugm.user_id IN`, memberUserIDs,
	)
	groupRows, err := r.db.Query(groupQuery, groupArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query team groups: %w", err)
	}
	defer groupRows.Close()

	for groupRows.Next() {
		group := &domain.UserGroup{}
		var description, icon sql.NullString
		if err := groupRows.Scan(
			&group.ID,
			&group.Name,
			&description,
			&group.AllUsersGroup,
			&icon,
			&group.IsUpdatable,
			&group.IsRemovable,
		); err != nil {
			return nil, fmt.Errorf("failed to scan team group: %w", err)
		}
		if description.Valid {
			group.Description = description.String
		}
		if icon.Valid {
			group.Icon = icon.String
		}
		group.Type = "UserGroup"
		people.GroupsInTeam = append(people.GroupsInTeam, group)
	}
	if err := groupRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team group rows: %w", err)
	}

	return people, nil
}

func (r *ProjectPeopleRepository) GetProjectDashboard(projectID string) (*domain.ProjectDashboard, error) {
	query := `
		SELECT pdw.id, pdw.widget_id, pdw.key, pdw.x, pdw.y, pdw.width, pdw.height, pdw.settings,
		       dw.name, dw.key, dw.app_id, dw.app_name, dw.extension_point,
		       dw.configurable, dw.collapsed, dw.borderless, dw.show_header,
		       dw.default_height, dw.default_width, dw.icon_path, dw.index_path,
		       dw.vendor_name, dw.vendor_url
		FROM project_dashboard_widgets pdw
		INNER JOIN dashboard_widgets dw ON pdw.widget_id = dw.id
		WHERE pdw.project_id = $1
		ORDER BY pdw.y, pdw.x
	`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query project dashboard widgets: %w", err)
	}
	defer rows.Close()

	dashboard := &domain.ProjectDashboard{
		Type: "ProjectDashboard",
	}

	for rows.Next() {
		var (
			id, widgetID                 string
			key, settings                sql.NullString
			x, y, width, height         int
			wName, wKey                  string
			appID, appName               sql.NullString
			extensionPoint               sql.NullString
			configurable, collapsed      bool
			borderless, showHeader       bool
			defaultHeight, defaultWidth  sql.NullString
			iconPath, indexPath          sql.NullString
			vendorName, vendorURL        sql.NullString
		)

		if err := rows.Scan(
			&id, &widgetID, &key, &x, &y, &width, &height, &settings,
			&wName, &wKey, &appID, &appName, &extensionPoint,
			&configurable, &collapsed, &borderless, &showHeader,
			&defaultHeight, &defaultWidth, &iconPath, &indexPath,
			&vendorName, &vendorURL,
		); err != nil {
			return nil, fmt.Errorf("failed to scan dashboard widget: %w", err)
		}

		embedding := &domain.DashboardWidgetEmbedding{
			ID:    id,
			Key:   key.String,
			X:     x,
			Y:     y,
			Width: width,
			Height: height,
			Widget: &domain.WidgetRef{
				ID:   widgetID,
				Type: "Widget",
			},
			Settings: settings.String,
			Type:     "DashboardWidgetEmbedding",
		}

		dashboard.Widgets = append(dashboard.Widgets, embedding)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating dashboard widget rows: %w", err)
	}

	return dashboard, nil
}

func buildInClause(prefix string, values []string) (string, []interface{}) {
	placeholders := make([]string, len(values))
	args := make([]interface{}, len(values))
	for i, v := range values {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = v
	}
	query := fmt.Sprintf("%s (%s)", prefix, strings.Join(placeholders, ", "))
	return query, args
}
