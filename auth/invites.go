package auth

import (
	"log"
	"time"

	"lod2/db"

	"go.jetify.com/typeid"
)

// authInvites table has rows:
// inviteId TEXT -- the unique invite code
// userId TEXT -- the user who created this invite
// issuedAt INTEGER - when the invite was created
// inviteLimit INTEGER - how many times this invite can be used

// Creates a new invite for the given user ID; returns the invite ID if successful, error otherwise.
func adminCreateInvite(userId string, inviteLimit int) (string, error) {
	inviteId, _ := typeid.WithPrefix("inv")

	_, err := db.DB.Exec("INSERT INTO authInvites (inviteId, userId, issuedAt, inviteLimit) VALUES (?, ?, ?, ?)", inviteId, userId, time.Now().Unix(), inviteLimit)

	if err != nil {
		log.Println("error creating invite:", err)
		return "", err
	}

	return inviteId.String(), nil
}

// Resets the specified user's invite to have at least this many invites left.
func AdminSetRemainingInvites(userId string, remainingInvites int) error {
	newInviteId, _ := typeid.WithPrefix("inv")

	currentInvitesConsumed, err := invitesConsumed(userId)

	log.Printf("currentInvitesConsumed: %d", currentInvitesConsumed)

	if err != nil {
		log.Println("error selecting remaining invites:", err)
		return err
	}

	newInviteLimit := currentInvitesConsumed + remainingInvites

	_, err = db.DB.Exec(`
    INSERT INTO authInvites (inviteId, userId, issuedAt, inviteLimit)
    VALUES (?, ?, ?, ?)
    ON CONFLICT(inviteId)
    DO UPDATE SET inviteLimit = excluded.inviteLimit
    ON CONFLICT(userId)
    DO UPDATE SET inviteLimit = excluded.inviteLimit
`, newInviteId, userId, time.Now().Unix(), newInviteLimit)

	if err != nil {
		log.Println("error updating invite:", err)
		return err
	}

	return nil
}

func invitesConsumed(userId string) (int, error) {
	var invitesConsumed int

	err := db.DB.QueryRow(`
		SELECT COUNT(authUsers.userId) as invitesConsumed
		FROM authInvites
		LEFT JOIN authUsers ON authInvites.inviteId = authUsers.inviteId
		WHERE authInvites.userId = ?`, userId).Scan(&invitesConsumed)

	if err != nil {
		log.Println("error selecting consumed invites:", err)
		return 0, err
	}

	return invitesConsumed, nil
}

func invitesRemaining(userId string) (int, error) {
	var invitesRemaining int

	err := db.DB.QueryRow(`
		SELECT SUM(authInvites.inviteLimit) - COUNT(authUsers.userId) as invitesRemaining
		FROM authInvites
		LEFT JOIN authUsers ON authInvites.inviteId = authUsers.inviteId
		WHERE authInvites.userId = ?`, userId).Scan(&invitesRemaining)

	if err != nil {
		log.Println("error selecting remaining invites:", err)
		return 0, err
	}

	return invitesRemaining, nil
}
