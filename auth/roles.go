package auth

import "database/sql"

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

var AllRoles = []Role{
	{Level: AccessLevelEdit, Scope: AccessScopeUserManagement},
	{Level: AccessLevelView, Scope: AccessScopeUserManagement},
	{Level: AccessLevelEdit, Scope: AccessScopeDangerousSql},
}

func GetRoleName(r Role) string {
	scope := ""
	switch r.Scope {
	case AccessScopeUserManagement:
		scope = "User Management"
	case AccessScopeDangerousSql:
		scope = "Dangerous SQL"
	default:
		scope = "(unknown scope)"
	}

	level := ""
	switch r.Level {
	case AccessLevelEdit:
		level = "Edit"
	case AccessLevelView:
		level = "View"
	case AccessLevelNone:
		level = "None"
	default:
		level = "(unknown level)"
	}

	return scope + ": " + level
}

// authRoles contains
// userId - foreign key to authUsers
// level - the access level
// scope - the access scope
func addRoles(tx *sql.Tx, userId string, roles []Role) error {
	for _, role := range roles {
		if _, err := tx.Exec(`INSERT INTO authRoles (userId, level, scope) VALUES (?, ?, ?)`, userId, role.Level, role.Scope); err != nil {
			return err
		}
	}
	return nil
}
