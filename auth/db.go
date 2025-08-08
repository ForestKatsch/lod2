package auth

import (
	"database/sql"
	"lod2/db"
	"log"
)

func migrate() {
	db.Migrate("auth", migrateAuth)
}

func migrateAuth(tx *sql.Tx, version int) (int, error) {
	// 1: delete existing authUsers, authSessions, authInvites tables if possible
	if version < 1 {
		tx.Exec(`DELETE TABLE authUsers`)
		tx.Exec(`DELETE TABLE authSessions`)
		tx.Exec(`DELETE TABLE authInvites`)

		version = 1
	}

	// 2: all previous migrations squashed
	if version < 2 {
		tx.Exec(`
			CREATE TABLE authInvites (
				inviteId TEXT PRIMARY KEY NOT NULL UNIQUE,
				userId TEXT NOT NULL UNIQUE,
				issuedAt INTEGER NOT NULL,
				inviteLimit INTEGER NOT NULL
			) WITHOUT ROWID`)

		tx.Exec(`
			CREATE TABLE authUsers (
				userId TEXT PRIMARY KEY NOT NULL UNIQUE,
				userName TEXT NOT NULL UNIQUE,
				userPasswordHash TEXT NOT NULL,
				inviteId TEXT DEFAULT NULL
			) WITHOUT ROWID`)

		tx.Exec(`
			CREATE TABLE authSessions (
				sessionId TEXT NOT NULL,
				userId TEXT NOT NULL,
				issuedAt INTEGER NOT NULL,
				refreshedAt INTEGER NOT NULL,
				expiresAt INTEGER NOT NULL,
		  	PRIMARY KEY (sessionId, userId)
			) WITHOUT ROWID`)

		version = 2
	}

	if version < 3 {
		tx.Exec(`
				CREATE TABLE authRoles (
					userId TEXT NOT NULL FOREIGN KEY REFERENCES authUsers(userId),
					group INTEGER NOT NULL,
				) WITHOUT ROWID`)

		version = 3
	}

	if version < 4 {
		tx.Exec("ALTER TABLE authRoles RENAME COLUMN group TO role")

		version = 4
	}

	if version < 5 {
		tx.Exec("ALTER TABLE authUsers ADD createdAt INTEGER NOT NULL DEFAULT 0")

		version = 5
	}

	// 6: refactor roles to level/scope instead of single role integer
	if version < 6 {
		// Create the new table schema
		tx.Exec(`
			CREATE TABLE authRoles_new (
				userId TEXT NOT NULL FOREIGN KEY REFERENCES authUsers(userId),
				level INTEGER NOT NULL,
				scope INTEGER NOT NULL
			) WITHOUT ROWID`)

		// Migrate old role values into level/scope pairs
		// Old mapping:
		//  0 (RoleUserEdit)   -> level=2 (Edit), scope=0 (UserManagement)
		//  1 (RoleUserView)   -> level=1 (View), scope=0 (UserManagement)
		//  2 (RoleDirectSql)  -> level=2 (Edit), scope=1 (DangerousSql)
		tx.Exec(`
			INSERT INTO authRoles_new (userId, level, scope)
			SELECT
				userId,
				CASE role WHEN 0 THEN 2 WHEN 1 THEN 1 WHEN 2 THEN 2 ELSE 0 END AS level,
				CASE role WHEN 0 THEN 0 WHEN 1 THEN 0 WHEN 2 THEN 1 ELSE 0 END AS scope
			FROM authRoles`)

		// Replace old table
		tx.Exec(`DROP TABLE authRoles`)
		tx.Exec(`ALTER TABLE authRoles_new RENAME TO authRoles`)

		version = 6
	}

	userId, _ := createUser(tx, "admin", "admin", AllRoles)
	log.Printf("created admin user with ID %s", userId)

	// Sneakily always update the admin user to have all roles.
	addRoles(tx, userId, AllRoles)

	return version, nil
}
