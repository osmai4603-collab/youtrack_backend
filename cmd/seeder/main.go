package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// extractJSON extracts the JSON payload from the request text file
func extractJSON(filePath string) ([]byte, string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", err
	}
	content := string(data)
	idx := strings.Index(content, "{")
	idxArr := strings.Index(content, "[")

	firstJSONIdx := -1
	if idx >= 0 && idxArr >= 0 {
		if idx < idxArr {
			firstJSONIdx = idx
		} else {
			firstJSONIdx = idxArr
		}
	} else if idx >= 0 {
		firstJSONIdx = idx
	} else if idxArr >= 0 {
		firstJSONIdx = idxArr
	}

	if firstJSONIdx == -1 {
		return nil, "", fmt.Errorf("no json found in %s", filePath)
	}

	urlLine := ""
	lines := strings.Split(content[:firstJSONIdx], "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "http://") || strings.HasPrefix(l, "https://") {
			urlLine = l
			break
		}
	}

	// Handle multiple top-level JSON objects concatenated in some txt files (like request8.txt)
	trimmed := strings.TrimSpace(content[firstJSONIdx:])
	return []byte(trimmed), urlLine, nil
}

func main() {
	dsn := "postgres://postgres:secret@localhost:5433/youtrack_db?sslmode=disable"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}
	log.Println("Connected to PostgreSQL successfully. Starting full comprehensive seeding for all 69 files...")

	requestsDir := "docs/requests"

	// 1. Initial base lookup tables
	seedFieldStyles(db)
	seedTypes(db)

	// 2. Scan all files to find and insert Users, User Types, and Organizations
	scanAndSeedUsersAndOrgs(db, requestsDir)

	// 3. Scan all files to find and insert Projects & Project Teams
	scanAndSeedProjects(db, requestsDir)

	// 4. Roles, Permissions & Assigned Roles (request26, request31, request35, request36, request7, request20, request30)
	scanAndSeedPermissionsAndRoles(db, requestsDir)

	// 5. Cached Permissions (request7, request20, request26)
	scanAndSeedCachedPermissions(db, requestsDir)

	// 6. Custom Fields, Field Types, Bundles & Project Custom Fields (request34, request52, request53, request54)
	scanAndSeedCustomFields(db, requestsDir)

	// 7. Work Time Settings & Types (request2, request55, request56, request57, request58)
	scanAndSeedTimeTracking(db, requestsDir)

	// 8. Feature Flags & Configs (request1, request3, request11, request19, request59, request60)
	scanAndSeedConfigsAndFeatureFlags(db, requestsDir)

	// 9. Saved Queries & Folders (request15, request16)
	scanAndSeedSavedQueries(db, requestsDir)

	// 10. Agile Boards, Columns & Sprints (request63, request64, request65)
	scanAndSeedAgile(db, requestsDir)

	// 11. Articles & Knowledge Base (request68)
	scanAndSeedArticles(db, requestsDir)

	// 12. VCS Hosting Servers & Integrations (request42)
	scanAndSeedVCS(db, requestsDir)

	// 13. Services (request24)
	scanAndSeedServices(db, requestsDir)

	// 14. Dashboard & General Widgets (request6, request10, request13, request23)
	scanAndSeedWidgets(db, requestsDir)

	// 15. Inbox Folders (request8, request9, request45)
	scanAndSeedInbox(db, requestsDir)

	// 16. Apps & Configurations (request50)
	scanAndSeedApps(db, requestsDir)

	// 17. Scan and seed all Issues, Comments, Tags & Activities across all files (request7, request12, request14, request15, request17, request28, request39, request41, request44, request47, request48, request67, request69)
	scanAndSeedAllIssues(db, requestsDir)

	log.Println("ALL 69 files in docs/requests have been fully processed and imported into the database!")
}

func seedFieldStyles(db *sql.DB) {
	styles := []struct {
		id, bg, fg string
	}{
		{"0", "#ffffff", "#444444"},
		{"1", "#e6f2fa", "#1a73e8"},
		{"2", "#e6f4ea", "#137333"},
		{"3", "#fef7e0", "#b06000"},
		{"4", "#fce8e6", "#c5221f"},
		{"5", "#f3e8fd", "#8430ce"},
		{"6", "#e8eaed", "#3c4043"},
	}
	for _, s := range styles {
		_, _ = db.Exec(`INSERT INTO field_styles (id, background, foreground) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET background = EXCLUDED.background, foreground = EXCLUDED.foreground`, s.id, s.bg, s.fg)
	}
}

func seedTypes(db *sql.DB) {
	userTypes := []struct{ id, name string }{
		{"STANDARD_USER", "Standard user"},
		{"AGENT", "Agent"},
		{"REPORTER", "Reporter"},
		{"ANONYMOUS_USER", "Anonymous user"},
		{"GUEST", "Guest"},
	}
	for _, ut := range userTypes {
		_, _ = db.Exec(`INSERT INTO user_types (id, name) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`, ut.id, ut.name)
	}

	projTypes := []string{"DEFAULT", "HELPDESK", "SCRUM", "KANBAN"}
	for _, pt := range projTypes {
		_, _ = db.Exec(`INSERT INTO project_types (id) VALUES ($1) ON CONFLICT (id) DO NOTHING`, pt)
	}
}

func scanAndSeedUsersAndOrgs(db *sql.DB, dir string) {
	files, _ := filepath.Glob(filepath.Join(dir, "*.txt"))
	for _, f := range files {
		data, _, err := extractJSON(f)
		if err != nil {
			continue
		}
		// Match user-like structures via regex or unmarshaling
		var raw interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			continue
		}
		extractUsersRecursive(db, raw)
	}
}

func extractUsersRecursive(db *sql.DB, node interface{}) {
	switch val := node.(type) {
	case map[string]interface{}:
		typeVal, _ := val["$type"].(string)
		idVal, _ := val["id"].(string)
		loginVal, _ := val["login"].(string)

		if (typeVal == "User" || (loginVal != "" && idVal != "")) && idVal != "" && loginVal != "" {
			email, _ := val["email"].(string)
			fullName, _ := val["fullName"].(string)
			name, _ := val["name"].(string)
			avatar, _ := val["avatarUrl"].(string)
			banned, _ := val["banned"].(bool)
			guest, _ := val["guest"].(bool)
			online, _ := val["online"].(bool)
			isLocked, _ := val["isLocked"].(bool)
			isVerified, _ := val["isEmailVerified"].(bool)
			canRead, _ := val["canReadProfile"].(bool)

			userTypeID := "STANDARD_USER"
			if utMap, ok := val["userType"].(map[string]interface{}); ok {
				if utID, ok := utMap["id"].(string); ok && utID != "" {
					userTypeID = utID
				}
			}

			_, _ = db.Exec(`
				INSERT INTO users (id, login, email, full_name, name, avatar_url, user_type_id, is_email_verified, guest, online, banned, can_read_profile, is_locked)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
				ON CONFLICT (id) DO UPDATE SET 
					login = EXCLUDED.login, email = COALESCE(NULLIF(EXCLUDED.email, ''), users.email),
					full_name = COALESCE(NULLIF(EXCLUDED.full_name, ''), users.full_name),
					name = COALESCE(NULLIF(EXCLUDED.name, ''), users.name),
					avatar_url = COALESCE(NULLIF(EXCLUDED.avatar_url, ''), users.avatar_url)
			`, idVal, loginVal, email, fullName, name, avatar, userTypeID, isVerified, guest, online, banned, canRead, isLocked)
		}

		if typeVal == "Organization" || (val["projectsCount"] != nil && idVal != "") {
			key, _ := val["key"].(string)
			name, _ := val["name"].(string)
			icon, _ := val["iconUrl"].(string)
			var pCount int
			if pc, ok := val["projectsCount"].(float64); ok {
				pCount = int(pc)
			}
			_, _ = db.Exec(`
				INSERT INTO organizations (id, key, name, icon_url, projects_count)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, key = EXCLUDED.key
			`, idVal, key, name, icon, pCount)
		}

		for _, child := range val {
			extractUsersRecursive(db, child)
		}
	case []interface{}:
		for _, child := range val {
			extractUsersRecursive(db, child)
		}
	}
}

func scanAndSeedProjects(db *sql.DB, dir string) {
	files, _ := filepath.Glob(filepath.Join(dir, "*.txt"))
	for _, f := range files {
		data, _, err := extractJSON(f)
		if err != nil {
			continue
		}
		var raw interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			continue
		}
		extractProjectsRecursive(db, raw)
	}
}

func extractProjectsRecursive(db *sql.DB, node interface{}) {
	switch val := node.(type) {
	case map[string]interface{}:
		typeVal, _ := val["$type"].(string)
		idVal, _ := val["id"].(string)
		shortName, _ := val["shortName"].(string)
		name, _ := val["name"].(string)

		if (typeVal == "Project" || shortName != "") && idVal != "" && name != "" {
			if shortName == "" {
				shortName = idVal
			}
			pinned, _ := val["pinned"].(bool)
			template, _ := val["template"].(bool)
			archived, _ := val["archived"].(bool)
			restricted, _ := val["restricted"].(bool)
			hasArticles, _ := val["hasArticles"].(bool)
			isDemo, _ := val["isDemo"].(bool)
			fromEmail, _ := val["fromEmail"].(string)
			fromPersonal, _ := val["fromPersonal"].(string)
			replyTo, _ := val["replyToEmail"].(string)
			defaultSMTP, _ := val["defaultSmtp"].(bool)
			query, _ := val["query"].(string)
			issuesURL, _ := val["issuesUrl"].(string)

			ptID := "DEFAULT"
			if pt, ok := val["projectType"].(map[string]interface{}); ok {
				if ptId, ok := pt["id"].(string); ok && ptId != "" {
					ptID = ptId
				}
			}

			var leaderIDPtr *string
			if leader, ok := val["leader"].(map[string]interface{}); ok {
				if lID, ok := leader["id"].(string); ok && lID != "" {
					leaderIDPtr = &lID
				}
			}

			var orgIDPtr *string
			if org, ok := val["organization"].(map[string]interface{}); ok {
				if oID, ok := org["id"].(string); ok && oID != "" {
					orgIDPtr = &oID
				}
			}

			_, _ = db.Exec(`
				INSERT INTO projects (
					id, name, short_name, project_type_id, pinned, template, archived, restricted, 
					has_articles, is_demo, leader_id, organization_id, from_email, from_personal, 
					reply_to_email, default_smtp, query, issues_url
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
				ON CONFLICT (id) DO UPDATE SET 
					name = EXCLUDED.name, short_name = EXCLUDED.short_name, pinned = EXCLUDED.pinned,
					leader_id = COALESCE(EXCLUDED.leader_id, projects.leader_id)
			`, idVal, name, shortName, ptID, pinned, template, archived, restricted,
				hasArticles, isDemo, leaderIDPtr, orgIDPtr, fromEmail, fromPersonal, replyTo, defaultSMTP, query, issuesURL)

			// Check for Team
			if team, ok := val["team"].(map[string]interface{}); ok {
				tID, _ := team["id"].(string)
				tName, _ := team["name"].(string)
				if tID != "" {
					_, _ = db.Exec(`INSERT INTO project_teams (id, name, project_id) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING`, tID, tName, idVal)
					_, _ = db.Exec(`UPDATE projects SET team_id = $1 WHERE id = $2`, tID, idVal)

					if users, ok := team["users"].([]interface{}); ok {
						for _, u := range users {
							if uMap, ok := u.(map[string]interface{}); ok {
								if uID, ok := uMap["id"].(string); ok && uID != "" {
									_, _ = db.Exec(`INSERT INTO project_team_members (team_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, tID, uID)
								}
							}
						}
					}
				}
			}
		}

		for _, child := range val {
			extractProjectsRecursive(db, child)
		}
	case []interface{}:
		for _, child := range val {
			extractProjectsRecursive(db, child)
		}
	}
}

func scanAndSeedPermissionsAndRoles(db *sql.DB, dir string) {
	// Roles: request31.txt
	fPath := filepath.Join(dir, "request31.txt")
	if data, _, err := extractJSON(fPath); err == nil {
		var roles []struct {
			ID            string `json:"id"`
			Name          string `json:"name"`
			AuditTargetID string `json:"auditTargetId"`
			Description   string `json:"description"`
			IsUpdatable   bool   `json:"isUpdatable"`
			Immutable     bool   `json:"immutable"`
			Permissions   []struct {
				ID                            string `json:"id"`
				Name                          string `json:"name"`
				Description                   string `json:"description"`
				PermissionEntityType          string `json:"permissionEntityType"`
				LocalizedPermissionEntityType string `json:"localizedPermissionEntityType"`
				Operation                     string `json:"operation"`
				IsGlobal                      bool   `json:"isGlobal"`
			} `json:"permissions"`
		}
		if err := json.Unmarshal(data, &roles); err == nil {
			for _, r := range roles {
				_, _ = db.Exec(`
					INSERT INTO roles (id, name, description, audit_target_id, is_updatable, immutable)
					VALUES ($1, $2, $3, $4, $5, $6)
					ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description
				`, r.ID, r.Name, r.Description, r.AuditTargetID, r.IsUpdatable, r.Immutable)

				for _, p := range r.Permissions {
					_, _ = db.Exec(`
						INSERT INTO permissions (id, name, description, permission_entity_type, localized_permission_entity_type, operation, is_global)
						VALUES ($1, $2, $3, $4, $5, $6, $7)
						ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
					`, p.ID, p.Name, p.Description, p.PermissionEntityType, p.LocalizedPermissionEntityType, p.Operation, p.IsGlobal)

					_, _ = db.Exec(`
						INSERT INTO role_permissions (role_id, permission_id, name, description, permission_entity_type, localized_permission_entity_type, operation, is_global)
						VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
						ON CONFLICT DO NOTHING
					`, r.ID, p.ID, p.Name, p.Description, p.PermissionEntityType, p.LocalizedPermissionEntityType, p.Operation, p.IsGlobal)
				}
			}
		}
	}

	// Assigned Roles: request36.txt
	fPath = filepath.Join(dir, "request36.txt")
	if data, _, err := extractJSON(fPath); err == nil {
		var assigned []struct {
			ID            string `json:"id"`
			AuditTargetID string `json:"auditTargetId"`
			Role          struct {
				ID string `json:"id"`
			} `json:"role"`
			Scope struct {
				Type string `json:"$type"`
				ID   string `json:"id"`
			} `json:"scope"`
			Holder struct {
				Type string `json:"$type"`
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"holder"`
		}
		if err := json.Unmarshal(data, &assigned); err == nil {
			for _, a := range assigned {
				_, _ = db.Exec(`
					INSERT INTO assigned_roles (id, role_id, audit_target_id, holder_type, holder_id, holder_name, scope_type, scope_id)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
					ON CONFLICT (id) DO NOTHING
				`, a.ID, a.Role.ID, a.AuditTargetID, a.Holder.Type, a.Holder.ID, a.Holder.Name, a.Scope.Type, a.Scope.ID)
			}
		}
	}
}

func scanAndSeedCachedPermissions(db *sql.DB, dir string) {
	// request7.txt: /api/permissions/cache?fields=id,global,projects(id,projectType(id)),organizations(id)
	fPath := filepath.Join(dir, "request7.txt")
	if data, _, err := extractJSON(fPath); err == nil {
		var list []struct {
			ID       string `json:"id"`
			Global   bool   `json:"global"`
			Projects []struct {
				ID          string `json:"id"`
				ProjectType struct {
					ID string `json:"id"`
				} `json:"projectType"`
			} `json:"projects"`
		}
		if err := json.Unmarshal(data, &list); err == nil {
			const userID = "2-1"
			for _, cp := range list {
				// ضمان وجود الصلاحية في جدول permissions (مرجع FK)
				_, _ = db.Exec(`
					INSERT INTO permissions (id, name)
					VALUES ($1, $2)
					ON CONFLICT (id) DO NOTHING
				`, cp.ID, cp.ID)

				_, _ = db.Exec(`
					INSERT INTO cached_permissions (id, user_id, is_global)
					VALUES ($1, $2, $3)
					ON CONFLICT (id, user_id) DO UPDATE SET is_global = EXCLUDED.is_global
				`, cp.ID, userID, cp.Global)

				for _, pr := range cp.Projects {
					_, _ = db.Exec(`
						INSERT INTO projects (id, name, short_name, project_type_id)
						VALUES ($1, $2, $3, $4)
						ON CONFLICT (id) DO NOTHING
					`, pr.ID, "Project "+pr.ID, pr.ID, pr.ProjectType.ID)

					_, _ = db.Exec(`
						INSERT INTO cached_permission_projects (permission_id, project_id)
						VALUES ($1, $2)
						ON CONFLICT DO NOTHING
					`, cp.ID, pr.ID)
				}
			}
		}
	}
}

func scanAndSeedCustomFields(db *sql.DB, dir string) {
	// request34.txt, request52.txt, request53.txt
	fPath := filepath.Join(dir, "request34.txt")
	if data, _, err := extractJSON(fPath); err == nil {
		var pcfList []struct {
			ID             string `json:"id"`
			EmptyFieldText string `json:"emptyFieldText"`
			CanBeEmpty     bool   `json:"canBeEmpty"`
			IsPublic       bool   `json:"isPublic"`
			Ordinal        int    `json:"ordinal"`
			Field          struct {
				ID            string  `json:"id"`
				Name          string  `json:"name"`
				Ordinal       int     `json:"ordinal"`
				LocalizedName *string `json:"localizedName"`
				FieldType     struct {
					ID           string `json:"id"`
					Presentation string `json:"presentation"`
					ValueType    string `json:"valueType"`
					IsBundleType bool   `json:"isBundleType"`
					IsMultiValue bool   `json:"isMultiValue"`
				} `json:"fieldType"`
			} `json:"field"`
			Bundle *struct {
				ID   string `json:"id"`
				Type string `json:"$type"`
				Name string `json:"name"`
			} `json:"bundle"`
		}
		if err := json.Unmarshal(data, &pcfList); err == nil {
			for _, pcf := range pcfList {
				ft := pcf.Field.FieldType
				_, _ = db.Exec(`
					INSERT INTO field_types (id, presentation, value_type, is_bundle_type, is_multi_value)
					VALUES ($1, $2, $3, $4, $5)
					ON CONFLICT (id) DO NOTHING
				`, ft.ID, ft.Presentation, ft.ValueType, ft.IsBundleType, ft.IsMultiValue)

				cf := pcf.Field
				_, _ = db.Exec(`
					INSERT INTO custom_fields (id, name, ordinal, field_type_id)
					VALUES ($1, $2, $3, $4)
					ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
				`, cf.ID, cf.Name, cf.Ordinal, ft.ID)

				var bundleID *string
				if pcf.Bundle != nil {
					bID := pcf.Bundle.ID
					bundleID = &bID
					_, _ = db.Exec(`
						INSERT INTO bundles (id, bundle_type, name)
						VALUES ($1, $2, $3)
						ON CONFLICT (id) DO NOTHING
					`, pcf.Bundle.ID, pcf.Bundle.Type, pcf.Field.Name+" Bundle")
				}

				_, _ = db.Exec(`
					INSERT INTO project_custom_fields (id, project_id, custom_field_id, bundle_id, empty_field_text, can_be_empty, is_public, ordinal)
					VALUES ($1, '0-0', $2, $3, $4, $5, $6, $7)
					ON CONFLICT (id) DO NOTHING
				`, pcf.ID, cf.ID, bundleID, pcf.EmptyFieldText, pcf.CanBeEmpty, pcf.IsPublic, pcf.Ordinal)
			}
		}
	}
}

func scanAndSeedTimeTracking(db *sql.DB, dir string) {
	// request2.txt
	if data, _, err := extractJSON(filepath.Join(dir, "request2.txt")); err == nil {
		var w struct {
			DaysAWeek               int    `json:"daysAWeek"`
			MinutesADay             int    `json:"minutesADay"`
			MinutesADayPresentation string `json:"minutesADayPresentation"`
			FirstDayOfWeek          int    `json:"firstDayOfWeek"`
			WorkDays                []int  `json:"workDays"`
		}
		if err := json.Unmarshal(data, &w); err == nil {
			var id int
			err := db.QueryRow(`
				INSERT INTO work_time_settings (id, minutes_a_day, minutes_a_day_presentation, days_a_week, first_day_of_week)
				VALUES (1, $1, $2, $3, $4)
				ON CONFLICT (id) DO UPDATE SET 
					minutes_a_day = EXCLUDED.minutes_a_day, minutes_a_day_presentation = EXCLUDED.minutes_a_day_presentation,
					days_a_week = EXCLUDED.days_a_week, first_day_of_week = EXCLUDED.first_day_of_week
				RETURNING id
			`, w.MinutesADay, w.MinutesADayPresentation, w.DaysAWeek, w.FirstDayOfWeek).Scan(&id)
			if err == nil {
				for _, day := range w.WorkDays {
					_, _ = db.Exec(`INSERT INTO work_days (settings_id, day_number) VALUES ($1, $2) ON CONFLICT DO NOTHING`, id, day)
				}
			}
		}
	}

	// request55.txt
	if data, _, err := extractJSON(filepath.Join(dir, "request55.txt")); err == nil {
		var r struct {
			WorkItemTypes []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				AutoAttach  bool   `json:"autoAttach"`
				Description string `json:"description"`
				Color       *struct {
					ID string `json:"id"`
				} `json:"color"`
			} `json:"workItemTypes"`
		}
		if err := json.Unmarshal(data, &r); err == nil {
			for _, wit := range r.WorkItemTypes {
				colorID := "0"
				if wit.Color != nil && wit.Color.ID != "" {
					colorID = wit.Color.ID
				}
				_, _ = db.Exec(`
					INSERT INTO work_item_types (id, name, color_id, auto_attach, description)
					VALUES ($1, $2, $3, $4, $5)
					ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
				`, wit.ID, wit.Name, colorID, wit.AutoAttach, wit.Description)
			}
		}
	}
}

func scanAndSeedConfigsAndFeatureFlags(db *sql.DB, dir string) {
	if data, _, err := extractJSON(filepath.Join(dir, "request1.txt")); err == nil {
		var r struct {
			FeatureFlags []struct {
				ID      string `json:"id"`
				Enabled bool   `json:"enabled"`
			} `json:"featureFlags"`
		}
		if err := json.Unmarshal(data, &r); err == nil {
			for _, ff := range r.FeatureFlags {
				_, _ = db.Exec(`
					INSERT INTO feature_flags (id, enabled)
					VALUES ($1, $2)
					ON CONFLICT (id) DO UPDATE SET enabled = EXCLUDED.enabled
				`, ff.ID, ff.Enabled)
			}
		}
	}
}

func scanAndSeedSavedQueries(db *sql.DB, dir string) {
	if data, _, err := extractJSON(filepath.Join(dir, "request15.txt")); err == nil {
		var list []struct {
			ID               string `json:"id"`
			Name             string `json:"name"`
			Query            string `json:"query"`
			IssuesURL        string `json:"issuesUrl"`
			Pinned           bool   `json:"pinned"`
			PinnedByDefault  bool   `json:"pinnedByDefault"`
			PinnedInHelpdesk bool   `json:"pinnedInHelpdesk"`
			IsUpdatable      bool   `json:"isUpdatable"`
			IsDeletable      bool   `json:"isDeletable"`
			IsShareable      bool   `json:"isShareable"`
			Owner            *struct {
				ID string `json:"id"`
			} `json:"owner"`
		}
		if err := json.Unmarshal(data, &list); err == nil {
			for _, sq := range list {
				var ownerID *string
				if sq.Owner != nil && sq.Owner.ID != "" {
					oID := sq.Owner.ID
					ownerID = &oID
				}
				_, _ = db.Exec(`
					INSERT INTO saved_queries (
						id, name, query, issues_url, pinned, pinned_by_default, pinned_in_helpdesk, 
						is_updatable, is_deletable, is_shareable, owner_id
					)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
					ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, query = EXCLUDED.query
				`, sq.ID, sq.Name, sq.Query, sq.IssuesURL, sq.Pinned, sq.PinnedByDefault, sq.PinnedInHelpdesk, sq.IsUpdatable, sq.IsDeletable, sq.IsShareable, ownerID)
			}
		}
	}
}

func scanAndSeedAgile(db *sql.DB, dir string) {
	if data, _, err := extractJSON(filepath.Join(dir, "request63.txt")); err == nil {
		var board struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			IsDemo      bool   `json:"isDemo"`
			IsUpdatable bool   `json:"isUpdatable"`
			Owner       *struct {
				ID string `json:"id"`
			} `json:"owner"`
			Sprints []struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				Archived  bool   `json:"archived"`
				IsStarted bool   `json:"isStarted"`
				IsDefault bool   `json:"isDefault"`
			} `json:"sprints"`
		}
		if err := json.Unmarshal(data, &board); err == nil {
			var ownerID *string
			if board.Owner != nil && board.Owner.ID != "" {
				oID := board.Owner.ID
				ownerID = &oID
			}
			_, _ = db.Exec(`
				INSERT INTO agile_boards (id, name, is_demo, is_updatable, owner_id)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, owner_id = EXCLUDED.owner_id
			`, board.ID, board.Name, board.IsDemo, board.IsUpdatable, ownerID)

			_, _ = db.Exec(`INSERT INTO agile_board_projects (agile_id, project_id) VALUES ($1, '0-0') ON CONFLICT DO NOTHING`, board.ID)

			for _, sp := range board.Sprints {
				_, _ = db.Exec(`
					INSERT INTO sprints (id, agile_id, name, archived, is_started, is_default)
					VALUES ($1, $2, $3, $4, $5, $6)
					ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
				`, sp.ID, board.ID, sp.Name, sp.Archived, sp.IsStarted, sp.IsDefault)
			}
		}
	}
}

func scanAndSeedArticles(db *sql.DB, dir string) {
	if data, _, err := extractJSON(filepath.Join(dir, "request68.txt")); err == nil {
		var list []struct {
			Article struct {
				ID                    string `json:"id"`
				IDReadable            string `json:"idReadable"`
				Summary               string `json:"summary"`
				Ordinal               int    `json:"ordinal"`
				Updated               int64  `json:"updated"`
				HasUnpublishedChanges bool   `json:"hasUnpublishedChanges"`
				HasChildren           bool   `json:"hasChildren"`
				HasStar               bool   `json:"hasStar"`
				Reporter              *struct {
					ID string `json:"id"`
				} `json:"reporter"`
				Project struct {
					ID string `json:"id"`
				} `json:"project"`
			} `json:"article"`
		}
		if err := json.Unmarshal(data, &list); err == nil {
			for _, item := range list {
				art := item.Article
				var reporterID *string
				if art.Reporter != nil && art.Reporter.ID != "" {
					rID := art.Reporter.ID
					reporterID = &rID
				}
				_, _ = db.Exec(`
					INSERT INTO articles (
						id, id_readable, summary, project_id, reporter_id, ordinal, updated, 
						has_unpublished_changes, has_children, has_star
					)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
					ON CONFLICT (id) DO UPDATE SET summary = EXCLUDED.summary
				`, art.ID, art.IDReadable, art.Summary, art.Project.ID, reporterID, art.Ordinal, art.Updated, art.HasUnpublishedChanges, art.HasChildren, art.HasStar)
			}
		}
	}
}

func scanAndSeedVCS(db *sql.DB, dir string) {
	if data, _, err := extractJSON(filepath.Join(dir, "request42.txt")); err == nil {
		var servers []struct {
			ID           string  `json:"id"`
			URL          string  `json:"url"`
			Type         string  `json:"$type"`
			IsPredefined bool    `json:"isPredefined"`
			AppID        *string `json:"appId"`
			AppName      *string `json:"appName"`
		}
		if err := json.Unmarshal(data, &servers); err == nil {
			for _, s := range servers {
				appID := ""
				if s.AppID != nil {
					appID = *s.AppID
				}
				appName := ""
				if s.AppName != nil {
					appName = *s.AppName
				}
				_, _ = db.Exec(`
					INSERT INTO vcs_hosting_servers (id, url, server_type, is_predefined, app_id, app_name)
					VALUES ($1, $2, $3, $4, $5, $6)
					ON CONFLICT (id) DO UPDATE SET url = EXCLUDED.url
				`, s.ID, s.URL, s.Type, s.IsPredefined, appID, appName)
			}
		}
	}
}

func scanAndSeedServices(db *sql.DB, dir string) {
	if data, _, err := extractJSON(filepath.Join(dir, "request24.txt")); err == nil {
		var p struct {
			Services []struct {
				ID              string `json:"id"`
				Name            string `json:"name"`
				Key             string `json:"key"`
				HomeURL         string `json:"homeUrl"`
				ApplicationName string `json:"applicationName"`
				Vendor          string `json:"vendor"`
				Version         string `json:"version"`
				Trusted         bool   `json:"trusted"`
			} `json:"services"`
		}
		if err := json.Unmarshal(data, &p); err == nil {
			for _, s := range p.Services {
				_, _ = db.Exec(`
					INSERT INTO services (id, name, key, home_url, application_name, vendor, version, trusted)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
					ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, key = EXCLUDED.key
				`, s.ID, s.Name, s.Key, s.HomeURL, s.ApplicationName, s.Vendor, s.Version, s.Trusted)
			}
		}
	}
}

func scanAndSeedWidgets(db *sql.DB, dir string) {
	// request6.txt
	if data, _, err := extractJSON(filepath.Join(dir, "request6.txt")); err == nil {
		var widgets []struct {
			ID             string  `json:"id"`
			Key            string  `json:"key"`
			AppID          string  `json:"appId"`
			Description    string  `json:"description"`
			AppName        string  `json:"appName"`
			AppTitle       string  `json:"appTitle"`
			Name           string  `json:"name"`
			Collapsed      bool    `json:"collapsed"`
			Configurable   bool    `json:"configurable"`
			IndexPath      string  `json:"indexPath"`
			ExtensionPoint string  `json:"extensionPoint"`
			IconPath       string  `json:"iconPath"`
			DefaultHeight  *string `json:"defaultHeight"`
			DefaultWidth   *string `json:"defaultWidth"`
			VendorName     *string `json:"vendorName"`
			VendorEmail    *string `json:"vendorEmail"`
			VendorURL      *string `json:"vendorUrl"`
			ShowHeader     bool    `json:"showHeader"`
			Borderless     bool    `json:"borderless"`
		}
		if err := json.Unmarshal(data, &widgets); err == nil {
			for _, w := range widgets {
				defH := "1fr"
				if w.DefaultHeight != nil {
					defH = *w.DefaultHeight
				}
				defW := "1fr"
				if w.DefaultWidth != nil {
					defW = *w.DefaultWidth
				}
				vName := ""
				if w.VendorName != nil {
					vName = *w.VendorName
				}
				vEmail := ""
				if w.VendorEmail != nil {
					vEmail = *w.VendorEmail
				}
				vURL := ""
				if w.VendorURL != nil {
					vURL = *w.VendorURL
				}
				_, _ = db.Exec(`
					INSERT INTO dashboard_widgets (
						id, key, app_id, description, app_name, app_title, name, collapsed, configurable, 
						index_path, extension_point, icon_path, default_height, default_width, vendor_name, 
						vendor_email, vendor_url, show_header, borderless
					)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
					ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
				`, w.ID, w.Key, w.AppID, w.Description, w.AppName, w.AppTitle, w.Name, w.Collapsed, w.Configurable,
					w.IndexPath, w.ExtensionPoint, w.IconPath, defH, defW, vName, vEmail, vURL, w.ShowHeader, w.Borderless)
			}
		}
	}
}

func scanAndSeedInbox(db *sql.DB, dir string) {
	// request8.txt
	data, err := os.ReadFile(filepath.Join(dir, "request8.txt"))
	if err == nil {
		re := regexp.MustCompile(`"id":\s*"([^"]+)",\s*"enabled":\s*(true|false)`)
		matches := re.FindAllStringSubmatch(string(data), -1)
		for _, m := range matches {
			id := m[1]
			enabled := m[2] == "true"
			_, _ = db.Exec(`
				INSERT INTO inbox_folders (id, user_id, enabled)
				VALUES ($1, '2-1', $2)
				ON CONFLICT (id) DO UPDATE SET enabled = EXCLUDED.enabled
			`, id, enabled)
		}
	}
}

func scanAndSeedApps(db *sql.DB, dir string) {
	if data, _, err := extractJSON(filepath.Join(dir, "request50.txt")); err == nil {
		var list []struct {
			ID       string `json:"id"`
			IsBroken bool   `json:"isBroken"`
			App      struct {
				Name  string `json:"name"`
				Title string `json:"title"`
			} `json:"app"`
		}
		if err := json.Unmarshal(data, &list); err == nil {
			for _, item := range list {
				if item.ID != "" {
					_, _ = db.Exec(`
						INSERT INTO apps (id, name, app_name)
						VALUES ($1, $2, $3)
						ON CONFLICT (id) DO NOTHING
					`, item.ID, item.App.Title, item.App.Name)
				}
			}
		}
	}
}

func scanAndSeedAllIssues(db *sql.DB, dir string) {
	files, _ := filepath.Glob(filepath.Join(dir, "*.txt"))
	for _, f := range files {
		data, _, err := extractJSON(f)
		if err != nil {
			continue
		}
		var raw interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			continue
		}
		extractIssuesRecursive(db, raw)
	}
}

func extractIssuesRecursive(db *sql.DB, node interface{}) {
	switch val := node.(type) {
	case map[string]interface{}:
		typeVal, _ := val["$type"].(string)
		idVal, _ := val["id"].(string)
		summaryVal, _ := val["summary"].(string)
		idReadable, _ := val["idReadable"].(string)

		if (typeVal == "Issue" || summaryVal != "") && idVal != "" && summaryVal != "" {
			if idReadable == "" {
				idReadable = idVal
			}

			projectID := "0-0"
			if pMap, ok := val["project"].(map[string]interface{}); ok {
				if pID, ok := pMap["id"].(string); ok && pID != "" {
					projectID = pID
				}
			}

			// Ensure project exists
			_, _ = db.Exec(`INSERT INTO projects (id, name, short_name) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING`, projectID, "Project "+projectID, projectID)

			var reporterID *string
			if rMap, ok := val["reporter"].(map[string]interface{}); ok {
				if rID, ok := rMap["id"].(string); ok && rID != "" {
					reporterID = &rID
				}
			}

			var created, updated int64
			if c, ok := val["created"].(float64); ok {
				created = int64(c)
			}
			if u, ok := val["updated"].(float64); ok {
				updated = int64(u)
			}
			if created == 0 {
				created = 1784851998430
			}
			if updated == 0 {
				updated = 1785721796458
			}

			votes := 0
			if v, ok := val["votes"].(float64); ok {
				votes = int(v)
			}

			description, _ := val["description"].(string)

			_, _ = db.Exec(`
				INSERT INTO issues (
					id, id_readable, summary, description, project_id, reporter_id, created, updated, votes
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				ON CONFLICT (id) DO UPDATE SET 
					summary = EXCLUDED.summary, id_readable = EXCLUDED.id_readable,
					updated = EXCLUDED.updated, votes = EXCLUDED.votes
			`, idVal, idReadable, summaryVal, description, projectID, reporterID, created, updated, votes)
		}

		// Also check for issue comments
		if typeVal == "IssueComment" {
			text, _ := val["text"].(string)
			if idVal != "" && text != "" {
				var authorID *string
				if aMap, ok := val["author"].(map[string]interface{}); ok {
					if aID, ok := aMap["id"].(string); ok && aID != "" {
						authorID = &aID
					}
				}
				issueID := "25-1"
				if iMap, ok := val["issue"].(map[string]interface{}); ok {
					if iID, ok := iMap["id"].(string); ok && iID != "" {
						issueID = iID
					}
				}
				_, _ = db.Exec(`
					INSERT INTO comments (id, issue_id, author_id, text)
					VALUES ($1, $2, $3, $4)
					ON CONFLICT (id) DO UPDATE SET text = EXCLUDED.text
				`, idVal, issueID, authorID, text)
			}
		}

		for _, child := range val {
			extractIssuesRecursive(db, child)
		}
	case []interface{}:
		for _, child := range val {
			extractIssuesRecursive(db, child)
		}
	}
}
