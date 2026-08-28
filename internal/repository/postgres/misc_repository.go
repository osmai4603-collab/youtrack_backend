package postgres

import (
	"database/sql"
	"fmt"
	"youtrack_backend/internal/domain"
)

type VCSRepository struct {
	db *sql.DB
}

func NewVCSRepository(db *sql.DB) *VCSRepository {
	return &VCSRepository{db: db}
}

func (r *VCSRepository) GetVCSServers() ([]*domain.VCSServer, error) {
	rows, err := r.db.Query(`
		SELECT
			id,
			url,
			server_type,
			is_predefined,
			app_id,
			app_name,
			application_id,
			ssl_key_id
		FROM vcs_hosting_servers
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query vcs hosting servers: %w", err)
	}
	defer rows.Close()

	var servers []*domain.VCSServer
	for rows.Next() {
		var (
			id            string
			url           sql.NullString
			serverType    sql.NullString
			isPredefined  bool
			appID         sql.NullString
			appName       sql.NullString
			applicationID sql.NullString
			sslKeyID      sql.NullString
		)

		err := rows.Scan(
			&id,
			&url,
			&serverType,
			&isPredefined,
			&appID,
			&appName,
			&applicationID,
			&sslKeyID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vcs hosting server: %w", err)
		}

		server := &domain.VCSServer{
			ID:                id,
			IsPredefined:      isPredefined,
			ChangesProcessors: []any{},
		}

		if url.Valid {
			server.URL = url.String
		}

		if appID.Valid {
			server.AppID = &appID.String
		}

		if appName.Valid {
			server.AppName = appName.String
		}

		if applicationID.Valid {
			server.ApplicationID = &applicationID.String
		}

		if serverType.Valid {
			switch serverType.String {
			case "GitHub":
				server.Type = "GitHubServer"
			case "GitLab":
				server.Type = "GitLabServer"
			case "BitBucket":
				server.Type = "BitBucketServer"
			case "AzureDevOps":
				server.Type = "AzureReposServer"
			default:
				server.Type = "GenericVcsChangesServer"
			}
		} else {
			server.Type = "GenericVcsChangesServer"
		}

		if sslKeyID.Valid {
			server.SSLKey = &domain.SSLKey{
				ID:   sslKeyID.String,
				Type: "SSLKey",
			}
		}

		servers = append(servers, server)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating vcs hosting servers: %w", err)
	}

	return servers, nil
}

type AppRepository struct {
	db *sql.DB
}

func NewAppRepository(db *sql.DB) *AppRepository {
	return &AppRepository{db: db}
}

func (r *AppRepository) GetServicesPage() (*domain.ServicesPage, error) {
	return &domain.ServicesPage{
		ID:   "default",
		Type: "ServicesPage",
	}, nil
}

type HubRepository struct {
	db *sql.DB
}

func NewHubRepository(db *sql.DB) *HubRepository {
	return &HubRepository{db: db}
}

func (r *HubRepository) GetHubUser() (*domain.HubUser, error) {
	return nil, nil
}
