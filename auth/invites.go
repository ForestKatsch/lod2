package auth

import (
	"database/sql"
	"log"
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

// Returns any unused invite for the specified user, or error if none found
func GetUserInviteId(userId string) (string, error) {
	var inviteId string

	err := db.DB.QueryRow(`
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

// Marks an invite as consumed by a specific user
func AdminConsumeInvite(inviteId string, consumedByUserId string) error {
	_, err := db.DB.Exec(`
		UPDATE authInvites 
		SET consumedByUserId = ?, consumedAt = ?
		WHERE inviteId = ? AND consumedByUserId IS NULL`, 
		consumedByUserId, time.Now().Unix(), inviteId)
	
	return err
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
	// Check if user has UserManagement Edit permission (unlimited invites)
	roles, err := GetUserRoles(userId)
	if err == nil {
		roleMap := GetRoleMap(roles)
		if roleMap[UserManagement] >= Edit {
			return -1, nil // -1 indicates unlimited invites
		}
	}

	// Count unused invites for regular users
	var invitesRemaining int
	err = db.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM authInvites 
		WHERE createdByUserId = ? AND consumedByUserId IS NULL`, userId).Scan(&invitesRemaining)

	if err != nil {
		log.Println("error counting remaining invites:", err)
		return 0, err
	}

	return invitesRemaining, nil
}
