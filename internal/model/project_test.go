package model

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestProjectNormalize(t *testing.T) {
	p := &Project{
		ID:   "DEMO",
		Name: "Demo Project",
	}

	p.Normalize()

	assert.Equal(t, "Project", p.Type)

	p.Team = &ProjectTeamDetailed{ID: "team-1"}
	p.Normalize()
	assert.Equal(t, "ProjectTeam", p.Team.Type)

	p.Organization = &Organization{ID: "org-1"}
	p.Normalize()
	assert.Equal(t, "Organization", p.Organization.Type)
}
