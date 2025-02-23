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
// expiresAt INTEGER - after this time, the session is expired and the user must log in again
// issuedAt INTEGER - when the session was created

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

	// 3: Update expiresAt and issuedAt to be integers, not datetime
	if version < 3 {
		_, err := tx.Exec("ALTER TABLE authSessions ADD COLUMN expiresAtEpoch INTEGER DEFAULT 0")
		if err != nil {
			return version, err
		}

		_, err = tx.Exec("ALTER TABLE authSessions ADD COLUMN issuedAtEpoch INTEGER DEFAULT 0")
		if err != nil {
			return version, err
		}

		// Copy the data
		_, err = tx.Exec("UPDATE authSessions SET expiresAtEpoch = strftime('%s', expiresAt)")
		if err != nil {
			return version, err
		}

		_, err = tx.Exec("UPDATE authSessions SET issuedAtEpoch = strftime('%s', issuedAt)")
		if err != nil {
			return version, err
		}

		// Delete the old columns
		_, err = tx.Exec("ALTER TABLE authSessions DROP COLUMN expiresAt")
		if err != nil {
			return version, err
		}

		_, err = tx.Exec("ALTER TABLE authSessions DROP COLUMN issuedAt")
		if err != nil {
			return version, err
		}

		// Rename the columns
		_, err = tx.Exec("ALTER TABLE authSessions RENAME COLUMN expiresAtEpoch TO expiresAt")
		if err != nil {
			return version, err
		}

		_, err = tx.Exec("ALTER TABLE authSessions RENAME COLUMN issuedAtEpoch TO issuedAt")
		if err != nil {
			return version, err
		}

		version = 3
	}

	return version, nil
}

// Creates a new user session and returns the session ID.
func createUserSession(userId string) (string, error) {
	sessionId, _ := typeid.WithPrefix("session")
	expiresAt := time.Now().Add(RefreshTokenExpirationDuration)

	_, err := db.DB.Exec("INSERT INTO authSessions (sessionId, userId, issuedAt, expiresAt) VALUES (?, ?, ?, ?)", sessionId, userId, time.Now().Unix(), expiresAt.Unix())

	if err != nil {
		log.Println("error creating session:", err)
		return "", err
	}

	return sessionId.String(), nil
}

// Returns true if the session exists and is not expired.
func getUserSessionIsValid(sessionId string) bool {
	var expiresAt int64

	err := db.DB.QueryRow("SELECT expiresAt FROM authSessions WHERE sessionId = ? AND expiresAt > ?", sessionId, time.Now().Unix()).Scan(&expiresAt)

	if err != nil {
		log.Println("error getting session:", err)
		return false
	}

	return expiresAt > time.Now().Unix()
}

// Attempts to invalidate the user session by setting the expiration time to now.
func invalidateSession(sessionId string) error {
	_, err := db.DB.Exec("UPDATE authSessions SET expiresAt = ? WHERE sessionId = ?", time.Now().Unix(), sessionId)

	if err != nil {
		log.Println("error invalidating session:", err)
		return err
	}

	log.Printf("session %s invalidated", sessionId)

	return nil
}

// Attempts to invalidate all sessions for the provided user by setting the expiration time to now.
func AdminInvalidateAllSessions(userId string) error {
	_, err := db.DB.Exec("UPDATE authSessions SET expiresAt = ? WHERE userId = ?", time.Now().Unix(), userId)

	if err != nil {
		log.Printf("error invalidating sessions for %s: %v", userId, err)
		return err
	}

	log.Printf("all sessions for %s invalidated", userId)

	return nil
}
