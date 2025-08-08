package auth

import (
	"context"
	"database/sql"
	"lod2/db"
)

// AccessLevel indicates the action capability within a scope.
type AccessLevel int

const (
	// No access to this scope
	AccessLevelNone AccessLevel = iota
	// View-only access
	View
	// View and edit access
	Edit
)

var NameToAccessLevel = map[string]AccessLevel{
	"None": AccessLevelNone,
	"View": View,
	"Edit": Edit,
}

var AccessLevelToName = map[AccessLevel]string{
	AccessLevelNone: "None",
	View:            "View",
	Edit:            "Edit",
}

// AccessScope indicates what area the access applies to.
type AccessScope int

const (
	UserManagement AccessScope = iota
	DangerousSql
)

var NameToAccessScope = map[string]AccessScope{
	"UserManagement": UserManagement,
	"DangerousSql":   DangerousSql,
}

var AccessScopeToName = make(map[AccessScope]string)

func init() {
	for name, scope := range NameToAccessScope {
		AccessScopeToName[scope] = name
	}
}

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
	{Level: Edit, Scope: UserManagement},
	{Level: Edit, Scope: DangerousSql},
}

func GetScopeName(scope AccessScope) string {
	for name, s := range NameToAccessScope {
		if s == scope {
			return name
		}
	}

	return "(unknown scope)"
}

func GetLevelName(level AccessLevel) string {
	return AccessLevelToName[level]
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

// UserHasRole returns true if any role for the given scope meets or exceeds the provided minimum level.
func UserHasRole(roles []Role, scope AccessScope, minimumLevel AccessLevel) bool {
	for _, r := range roles {
		if r.Scope == scope && r.Level >= minimumLevel {
			return true
		}
	}
	return false
}

// VerifyRole loads the current user from context and checks if they have the specified role requirement.
// Returns true if authorized, false otherwise. If the user is not logged in, returns false.
func VerifyRole(ctx context.Context, scope AccessScope, minimumLevel AccessLevel) bool {
	userInfo := GetCurrentUserInfo(ctx)
	if userInfo == nil {
		return false
	}

	roles, err := GetUserRoles(userInfo.UserId)
	if err != nil {
		return false
	}
	return UserHasRole(roles, scope, minimumLevel)
}

func GetRoleMap(roles []Role) map[AccessScope]AccessLevel {
	roleMap := make(map[AccessScope]AccessLevel)
	for _, role := range roles {
		roleMap[role.Scope] = role.Level
	}
	return roleMap
}
