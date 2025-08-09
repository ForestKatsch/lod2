package auth

import (
	"database/sql"
	"lod2/db"
	"log"
	"time"

	"go.jetify.com/typeid"
)

func migrate() {
	db.Migrate("auth", migrateAuth)
}

func migrateAuth(tx *sql.Tx, version int) (int, error) {
	// 1: delete existing authUsers, authSessions, authInvites tables if possible
	if version < 1 {
		if _, err := tx.Exec(`DROP TABLE IF EXISTS authUsers`); err != nil {
			return version, err
		}
		if _, err := tx.Exec(`DROP TABLE IF EXISTS authSessions`); err != nil {
			return version, err
		}
		if _, err := tx.Exec(`DROP TABLE IF EXISTS authInvites`); err != nil {
			return version, err
		}

		version = 1
	}

	// 2: all previous migrations squashed
	if version < 2 {
		if _, err := tx.Exec(`
			CREATE TABLE authInvites (
				inviteId TEXT PRIMARY KEY NOT NULL UNIQUE,
				userId TEXT NOT NULL UNIQUE,
				issuedAt INTEGER NOT NULL,
				inviteLimit INTEGER NOT NULL
			) WITHOUT ROWID`); err != nil {
			return version, err
		}

		if _, err := tx.Exec(`
			CREATE TABLE authUsers (
				userId TEXT PRIMARY KEY NOT NULL UNIQUE,
				userName TEXT NOT NULL UNIQUE,
				userPasswordHash TEXT NOT NULL,
				inviteId TEXT DEFAULT NULL
			) WITHOUT ROWID`); err != nil {
			return version, err
		}

		if _, err := tx.Exec(`
			CREATE TABLE authSessions (
				sessionId TEXT NOT NULL,
				userId TEXT NOT NULL,
				issuedAt INTEGER NOT NULL,
				refreshedAt INTEGER NOT NULL,
				expiresAt INTEGER NOT NULL,
		  	PRIMARY KEY (sessionId, userId)
			) WITHOUT ROWID`); err != nil {
			return version, err
		}

		version = 2
	}

	if version < 5 {
		if _, err := tx.Exec("ALTER TABLE authUsers ADD createdAt INTEGER NOT NULL DEFAULT 0"); err != nil {
			return version, err
		}

		version = 5
	}

	// 7: refactor roles to level/scope instead of single role integer
	if version < 7 {
		if _, err := tx.Exec(`DROP TABLE IF EXISTS authRoles`); err != nil {
			return version, err
		}
		if _, err := tx.Exec(`
				CREATE TABLE authRoles (
					userId TEXT NOT NULL REFERENCES authUsers(userId),
					level INTEGER NOT NULL,
					scope INTEGER NOT NULL,
					PRIMARY KEY (userId, level, scope)
				) WITHOUT ROWID`); err != nil {
			return version, err
		}
		version = 7
	}

	if version < 8 {
		userId, err := createUser(tx, "admin", "admin", AllRoles)
		if err != nil {
			return version, err
		}
		log.Printf("created admin user with ID %s", userId)

		version = 8
	}

	if version < 9 {
		if _, err := tx.Exec("ALTER TABLE authUsers ADD COLUMN deleted INTEGER NOT NULL DEFAULT 0"); err != nil {
			return version, err
		}
		version = 9
	}

	if version < 10 {
		// Drop old invite system completely
		if _, err := tx.Exec("DROP TABLE IF EXISTS authInvites"); err != nil {
			return version, err
		}
		
		// Create new one-invite-per-row system  
		if _, err := tx.Exec(`
			CREATE TABLE authInvites (
				inviteId TEXT PRIMARY KEY NOT NULL UNIQUE,
				createdByUserId TEXT NOT NULL,
				consumedByUserId TEXT DEFAULT NULL,
				createdAt INTEGER NOT NULL,
				consumedAt INTEGER DEFAULT NULL
			) WITHOUT ROWID`); err != nil {
			return version, err
		}
		version = 10
	}

	// Sneakily always update the admin user to have all roles. This runs at every boot and makes sure the admin user has all roles even if additional scopes are added.
	userId, err := AdminGetUserIdByUsername(tx, "admin")
	if err != nil {
		return version, err
	}

	if err := addRoles(tx, userId, AllRoles); err != nil {
		return version, err
	}

	// Ensure admin has some invites (they have unlimited anyway due to UserManagement role, but for consistency)
	// Check if admin has any invites, if not give them 5
	var inviteCount int
	tx.QueryRow("SELECT COUNT(*) FROM authInvites WHERE createdByUserId = ? AND consumedByUserId IS NULL", userId).Scan(&inviteCount)
	if inviteCount == 0 {
		for i := 0; i < 5; i++ {
			inviteId, _ := typeid.WithPrefix("inv")
			tx.Exec("INSERT INTO authInvites (inviteId, createdByUserId, createdAt) VALUES (?, ?, ?)", 
				inviteId, userId, time.Now().Unix())
		}
	}

	return version, nil
}
