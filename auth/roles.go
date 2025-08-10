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
	"No access": AccessLevelNone,
	"View":      View,
	"Edit":      Edit,
}

var AccessLevelToName = make(map[AccessLevel]string)

var AllAccessLevels = []AccessLevel{
	AccessLevelNone,
	View,
	Edit,
}

// AccessScope indicates what area the access applies to.
type AccessScope int

const (
	UserManagement AccessScope = iota
	DangerousSql
	Storage
	Media
)

// This also defines the order of the scopes in the UI.
var AllAccessScopes = []AccessScope{
	DangerousSql,
	UserManagement,
	Storage,
	Media,
}

// These are the raw names that need to match enum names. Templates will need to be updated if these are changed.
var NameToAccessScope = map[string]AccessScope{
	"UserManagement": UserManagement,
	"DangerousSql":   DangerousSql,
	"Storage":        Storage,
	"Media":          Media,
}

var AccessScopeToDisplayName = map[AccessScope]string{
	UserManagement: "User management",
	DangerousSql:   "Database access",
	Storage:        "Files",
	Media:          "Media",
}

var AccessScopeToName = make(map[AccessScope]string)

func init() {
	for name, scope := range NameToAccessScope {
		AccessScopeToName[scope] = name
	}

	if len(AccessScopeToName) != len(AllAccessScopes) {
		panic("AccessScopeToName is not the same length as AllAccessScopes; make sure to add all scopes to NameToAccessScope")
	}

	if len(AccessScopeToDisplayName) != len(AllAccessScopes) {
		panic("AccessScopeToDisplayName is not the same length as AllAccessScopes; make sure to add all scopes to AccessScopeToDisplayName")
	}

	// Populate the reverse map from the source NameToAccessLevel map.
	for name, level := range NameToAccessLevel {
		AccessLevelToName[level] = name
	}

	if len(AccessLevelToName) != len(AllAccessLevels) {
		panic("AccessLevelToName is not the same length as AllAccessLevels; make sure to add all levels to NameToAccessLevel")
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
	{Level: Edit, Scope: Storage},
	{Level: Edit, Scope: Media},
}

func GetScopeName(scope AccessScope) string {
	return AccessScopeToName[scope]
}

func GetScopeDisplayName(scope AccessScope) string {
	return AccessScopeToDisplayName[scope]
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
func setRoles(tx *sql.Tx, userId string, roles []Role) error {
	for _, role := range roles {
		if role.Level == AccessLevelNone {
			if _, err := tx.Exec(`DELETE FROM authRoles WHERE userId = ? AND scope = ?`, userId, role.Scope); err != nil {
				return err
			}
		} else {
			if _, err := tx.Exec(`INSERT OR REPLACE INTO authRoles (userId, level, scope) VALUES (?, ?, ?)`, userId, role.Level, role.Scope); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetUserRoles returns roles for all scopes, with None for scopes not assigned to the user.
func GetUserRoles(userId string) ([]Role, error) {
	rows, err := db.DB.Query(`SELECT level, scope FROM authRoles WHERE userId = ?`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Build a map of existing roles from database
	existingRoles := make(map[AccessScope]AccessLevel)
	for rows.Next() {
		var level int
		var scope int
		if err := rows.Scan(&level, &scope); err != nil {
			return nil, err
		}
		existingRoles[AccessScope(scope)] = AccessLevel(level)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Build result with all scopes, using None for missing ones
	roles := make([]Role, 0, len(AllAccessScopes))
	for _, scope := range AllAccessScopes {
		level, exists := existingRoles[scope]
		if !exists {
			level = AccessLevelNone
		}
		roles = append(roles, Role{Level: level, Scope: scope})
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

func UserIsAdmin(ctx context.Context) bool {
	userInfo := GetCurrentUserInfo(ctx)
	if userInfo == nil {
		return false
	}

	if userInfo.Roles == nil {
		return false
	}

	if UserHasRole(userInfo.Roles, UserManagement, View) {
		return true
	}

	return false
}

func AdminSetUserRoles(userId string, roles []Role) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = setRoles(tx, userId, roles)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
