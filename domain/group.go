package domain

import "time"

// Group represents a user group backed by the `groups` table.
type Group struct {
	ID                   string     `json:"id" db:"id"`
	Name                 string     `json:"name" db:"name"`
	Description          *string    `json:"description" db:"description"`
	GroupType            string     `json:"groupType" db:"group_type"`
	ParentGroupID        *string    `json:"parentGroupId,omitempty" db:"parent_group_id"`
	UsersCount           int64      `json:"usersCount" db:"users_count"`
	Icon                 *string    `json:"icon,omitempty" db:"icon"`
	AllUsersGroup        bool       `json:"allUsersGroup" db:"all_users_group"`
	RegisteredUsersGroup bool       `json:"registeredUsersGroup" db:"registered_users_group"`
	AutoJoin             bool       `json:"autoJoin" db:"auto_join"`
	Users                []User     `json:"users" db:"-"`
	CreatedAt            *time.Time `json:"createdAt,omitempty" db:"created_at"`
	UpdatedAt            *time.Time `json:"updatedAt,omitempty" db:"updated_at"`
}
