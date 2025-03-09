package auth

import (
	"errors"
	"log"
	"time"

	"lod2/db"

	"go.jetify.com/typeid"
)

// authSessions table has rows:
// sessionId TEXT
// userId TEXT
// issuedAt INTEGER - when the session was created
// expiresAt INTEGER - after this time, the session is expired and the user must log in again
// refreshedAt INTEGER - the most recent time the access token was refreshed based on this session

// Creates a new user session and returns the session ID.
func createUserSession(userId string) (string, error) {
	sessionId, _ := typeid.WithPrefix("session")
	expiresAt := time.Now().Add(RefreshTokenExpirationDuration)

	_, err := db.DB.Exec("INSERT INTO authSessions (sessionId, userId, issuedAt, refreshedAt, expiresAt) VALUES (?, ?, ?, ?, ?)", sessionId, userId, time.Now().Unix(), time.Now().Unix(), expiresAt.Unix())

	if err != nil {
		log.Println("error creating session:", err)
		return "", err
	}

	return sessionId.String(), nil
}

// Updates the provided session to indicate it has just been refreshed.
func updateUserSessionRefresh(sessionId string) error {
	refreshedAt := time.Now().Unix()

	_, err := db.DB.Exec("UPDATE authSessions SET refreshedAt = ? WHERE sessionId = ?", refreshedAt, sessionId)

	if err != nil {
		log.Println("error updating session:", err)
		return err
	}

	return nil
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

type UserSession struct {
	SessionId   string
	IssuedAt    int64
	ExpiresAt   int64
	RefreshedAt int64
	Expired     bool
}

func AdminGetUserSessions(userId string) ([]UserSession, error) {
	rows, err := db.DB.Query(`
		SELECT
			sessionId,
			issuedAt,
			expiresAt,
			refreshedAt
		FROM authSessions
		WHERE userId = ?
		ORDER BY issuedAt DESC`, userId)

	if err != nil {
		return nil, errors.New("invalid user id")
	}

	defer rows.Close()

	var sessions []UserSession

	for rows.Next() {
		var session UserSession
		err := rows.Scan(&session.SessionId, &session.IssuedAt, &session.ExpiresAt, &session.RefreshedAt)
		if err != nil {
			return nil, err
		}

		session.Expired = session.ExpiresAt <= time.Now().Unix()

		sessions = append(sessions, session)
	}

	return sessions, nil
}
