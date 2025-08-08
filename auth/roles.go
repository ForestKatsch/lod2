package auth

import (
	"database/sql"
	"lod2/db"
)

// AccessLevel indicates the action capability within a scope.
type AccessLevel int

const (
	// No access to this scope
	AccessLevelNone AccessLevel = iota
	// View-only access
	AccessLevelView
	// View and edit access
	AccessLevelEdit
)

// AccessScope indicates what area the access applies to.
type AccessScope int

const (
	AccessScopeUserManagement AccessScope = iota
	AccessScopeDangerousSql
)

// Role is a combination of a level and a scope.
type Role struct {
	Level AccessLevel
	Scope AccessScope
}

type RoleString struct {
	Scope string
	Level string
}

var AllRoles = []Role{
	{Level: AccessLevelEdit, Scope: AccessScopeUserManagement},
	{Level: AccessLevelEdit, Scope: AccessScopeDangerousSql},
}

func GetScopeName(scope AccessScope) string {
	switch scope {
	case AccessScopeUserManagement:
		return "User Management"
	case AccessScopeDangerousSql:
		return "Dangerous SQL"
	default:
		return "(unknown scope)"
	}
}

func GetLevelName(level AccessLevel) string {
	switch level {
	case AccessLevelEdit:
		return "Edit"
	case AccessLevelView:
		return "View"
	case AccessLevelNone:
		return "None"
	default:
		return "(unknown level)"
	}
}

func GetRoleString(role Role) RoleString {
	return RoleString{
		Scope: GetScopeName(role.Scope),
		Level: GetLevelName(role.Level),
	}
}

func GetRoleStrings(roles []Role) []RoleString {
	roleStrings := make([]RoleString, 0, len(roles))
	for _, role := range roles {
		roleStrings = append(roleStrings, GetRoleString(role))
	}
	return roleStrings
}

// authRoles contains
// userId - foreign key to authUsers
// level - the access level
// scope - the access scope
func addRoles(tx *sql.Tx, userId string, roles []Role) error {
	for _, role := range roles {
		if _, err := tx.Exec(`INSERT OR REPLACE INTO authRoles (userId, level, scope) VALUES (?, ?, ?)`, userId, role.Level, role.Scope); err != nil {
			return err
		}
	}
	return nil
}

// GetUserRoles returns all roles assigned to the given user.
func GetUserRoles(userId string) ([]Role, error) {
	rows, err := db.DB.Query(`SELECT level, scope FROM authRoles WHERE userId = ?`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]Role, 0)
	for rows.Next() {
		var level int
		var scope int
		if err := rows.Scan(&level, &scope); err != nil {
			return nil, err
		}
		roles = append(roles, Role{Level: AccessLevel(level), Scope: AccessScope(scope)})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}
