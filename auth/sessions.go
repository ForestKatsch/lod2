package auth

import (
	"database/sql"
	"log"
	"time"

	"lod2/db"

	"go.jetify.com/typeid"
)

func migrateSessions() {
	db.MigrateTable("authSessions", migrateAuthSessionsTable)
}

// authSessions table has rows:
// sessionId TEXT
// userId TEXT
// expiresAt DATETIME - after this time, the session is expired and the user must log in again
// issuedAt DATETIME - when the session was created

func migrateAuthSessionsTable(tx *sql.Tx, version int) (int, error) {
	// 1: Table exists
	if version < 1 {
		_, err := tx.Exec("CREATE TABLE authSessions (sessionId TEXT PRIMARY KEY NOT NULL UNIQUE, userId TEXT NOT NULL, expiresAt DATETIME NOT NULL) WITHOUT ROWID")

		if err != nil {
			return version, err
		}

		version = 1
	}

	// 2: Add issuedAt
	if version < 2 {
		_, err := tx.Exec("ALTER TABLE authSessions ADD COLUMN issuedAt DATETIME DEFAULT CURRENT_TIMESTAMP")

		if err != nil {
			return version, err
		}

		version = 2
	}

	return version, nil
}

// Creates a new user session and returns the session ID.
func createUserSession(userId string) (string, error) {
	sessionId, _ := typeid.WithPrefix("session")
	expiresAt := time.Now().Add(RefreshTokenExpirationDuration)

	_, err := db.DB.Exec("INSERT INTO authSessions (sessionId, userId, issuedAt, expiresAt) VALUES (?, ?, ?, ?)", sessionId, userId, time.Now(), expiresAt)

	if err != nil {
		log.Println("error creating session:", err)
		return "", err
	}

	return sessionId.String(), nil
}

// Returns true if the session exists and is not expired.
func getUserSessionIsValid(sessionId string) bool {
	var expiresAt time.Time

	err := db.DB.QueryRow("SELECT expiresAt FROM authSessions WHERE sessionId = ? AND expiresAt > ?", sessionId, time.Now()).Scan(&expiresAt)

	if err != nil {
		log.Println("error getting session:", err)
		return false
	}

	return expiresAt.After(time.Now())
}

// Attempts to invalidate the user session by setting the expiration time to now.
func invalidateSession(sessionId string) error {
	_, err := db.DB.Exec("UPDATE authSessions SET expiresAt = ? WHERE sessionId = ?", time.Now(), sessionId)

	if err != nil {
		log.Println("error invalidating session:", err)
		return err
	}

	log.Printf("session %s invalidated", sessionId)

	return nil
}
