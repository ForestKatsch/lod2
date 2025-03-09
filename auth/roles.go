package auth

import "database/sql"

type Role int

const (
	RoleUserEdit Role = iota
	RoleUserView

	RoleDirectSql
)

var AllRoles = []Role{RoleUserEdit, RoleUserView, RoleDirectSql}

func GetRoleName(r Role) string {
	switch r {
	case RoleUserEdit:
		return "User Edit"
	case RoleUserView:
		return "User View"

	case RoleDirectSql:
		return "Direct SQL"

	default:
		return "(not a group)"
	}
}

// authRoles contains
// userId - foreign key to authUsers
// role - the role ID

func addRoles(tx *sql.Tx, userId string, roles []Role) error {
	for _, role := range roles {
		if _, err := tx.Exec(`INSERT INTO authRoles (userId, role) VALUES (?, ?)`, userId, role); err != nil {
			return err
		}
	}
	return nil
}
