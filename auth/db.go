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
				userId TEXT NOT NULL FOREIGN KEY REFERENCES authUsers(userId),
				issuedAt INTEGER NOT NULL,
				inviteLimit INTEGER NOT NULL
			) WITHOUT ROWID`)

		tx.Exec(`
			CREATE TABLE authUsers (
				userId TEXT PRIMARY KEY NOT NULL UNIQUE,
				userName TEXT NOT NULL UNIQUE,
				userPasswordHash TEXT NOT NULL UNIQUE,
				inviteId TEXT DEFAULT NULL
			) WITHOUT ROWID`)

		tx.Exec(`
			CREATE TABLE authSessions (
				sessionId TEXT PRIMARY KEY NOT NULL UNIQUE,
				userId TEXT NOT NULL FOREIGN KEY REFERENCES authUsers(userId),
				issuedAt INTEGER NOT NULL,
				refreshedAt INTEGER NOT NULL,
				expiresAt INTEGER NOT NULL
			) WITHOUT ROWID`)

		userId, _ := createUser(tx, "admin", "admin")
		log.Printf("created admin user with ID %s", userId)
		version = 2
	}

	return version, nil
}
