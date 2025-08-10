package auth

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"lod2/db"

	"go.jetify.com/typeid"
)

// authInvites table has rows:
// inviteId TEXT -- the unique invite code
// createdByUserId TEXT -- the user who created this invite
// consumedByUserId TEXT -- the user who used this invite (NULL if unused)
// createdAt INTEGER -- when the invite was created
// consumedAt INTEGER -- when the invite was consumed (NULL if unused)

// Creates a single new invite for the given user ID
func AdminCreateInvite(createdByUserId string) (string, error) {
	inviteId, _ := typeid.WithPrefix("inv")

	_, err := db.DB.Exec("INSERT INTO authInvites (inviteId, createdByUserId, createdAt) VALUES (?, ?, ?)",
		inviteId, createdByUserId, time.Now().Unix())

	if err != nil {
		log.Println("error creating invite:", err)
		return "", err
	}

	return inviteId.String(), nil
}

// Creates a single new invite for the given user ID within a transaction
func AdminCreateInviteTx(tx *sql.Tx, createdByUserId string) (string, error) {
	inviteId, _ := typeid.WithPrefix("inv")

	_, err := tx.Exec("INSERT INTO authInvites (inviteId, createdByUserId, createdAt) VALUES (?, ?, ?)",
		inviteId, createdByUserId, time.Now().Unix())

	if err != nil {
		log.Println("error creating invite:", err)
		return "", err
	}

	return inviteId.String(), nil
}

// Sets the user to have exactly this many unused invites
func AdminSetRemainingInvites(userId string, remainingInvites int) error {
	// Delete all unused invites for this user
	_, err := db.DB.Exec("DELETE FROM authInvites WHERE createdByUserId = ? AND consumedByUserId IS NULL", userId)
	if err != nil {
		return err
	}

	// Create the specified number of new invites
	for i := 0; i < remainingInvites; i++ {
		_, err := AdminCreateInvite(userId)
		if err != nil {
			return err
		}
	}

	return nil
}

// Returns the first unused invite for the specified user, or error if none found
func GetUserInviteId(userId string) (string, error) {
	var inviteId string

	err := db.DB.QueryRow(`
		SELECT inviteId
		FROM authInvites
		WHERE createdByUserId = ? AND consumedByUserId IS NULL
		ORDER BY inviteId DESC
		LIMIT 1`, userId).Scan(&inviteId)

	if err != nil {
		log.Println("error selecting unused invite code:", err)
		return "", err
	}

	return inviteId, nil
}

// Returns any unused invite for the specified user within a transaction
func GetUserInviteIdTx(tx *sql.Tx, userId string) (string, error) {
	var inviteId string

	err := tx.QueryRow(`
		SELECT inviteId
		FROM authInvites
		WHERE createdByUserId = ? AND consumedByUserId IS NULL
		LIMIT 1`, userId).Scan(&inviteId)

	if err != nil {
		log.Println("error selecting unused invite code:", err)
		return "", err
	}

	return inviteId, nil
}

// Marks an invite as consumed by a specific user within a transaction
func AdminConsumeInviteTx(tx *sql.Tx, inviteId string, consumedByUserId string) error {
	_, err := tx.Exec(`
		UPDATE authInvites 
		SET consumedByUserId = ?, consumedAt = ?
		WHERE inviteId = ? AND consumedByUserId IS NULL`,
		consumedByUserId, time.Now().Unix(), inviteId)

	return err
}

func AdminInvitesRemaining(userId string) (int, error) {
	// Count unused invites for all users
	var invitesRemaining int
	err := db.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM authInvites 
		WHERE createdByUserId = ? AND consumedByUserId IS NULL`, userId).Scan(&invitesRemaining)

	if err != nil {
		log.Println("error counting remaining invites:", err)
		return 0, err
	}

	return invitesRemaining, nil
}

// Validates an invite code and returns invite info if valid
func ValidateInviteCode(inviteCode string) (createdByUserId string, err error) {
	var createdBy string

	err = db.DB.QueryRow(`
		SELECT createdByUserId 
		FROM authInvites 
		WHERE inviteId = ? AND consumedByUserId IS NULL`, inviteCode).Scan(&createdBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("invalid or expired invite code")
		}
		return "", err
	}

	return createdBy, nil
}

// Registers a new user with an invite code
func RegisterUserWithInvite(inviteCode string, username string, password string) (string, error) {
	// Validate invite code first
	_, err := ValidateInviteCode(inviteCode)
	if err != nil {
		return "", err
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	// Create the new user using the existing helper function
	newUserId, err := createUserWithInvite(tx, username, password, inviteCode)
	if err != nil {
		return "", err
	}

	// Consume the invite
	err = AdminConsumeInviteTx(tx, inviteCode, newUserId)
	if err != nil {
		return "", err
	}

	// Give the new user 5 starting invites
	for i := 0; i < 5; i++ {
		_, err := AdminCreateInviteTx(tx, newUserId)
		if err != nil {
			return "", err
		}
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return newUserId, nil
}

// Generates a full invite URL for sharing
func GenerateInviteURL(hostname string, inviteCode string) string {
	// Use https for non-localhost hosts, http for localhost
	scheme := "https"
	if hostname == "localhost:10800" || strings.HasPrefix(hostname, "localhost") {
		scheme = "http"
	}
	return scheme + "://" + hostname + "/auth/invite/" + inviteCode
}
