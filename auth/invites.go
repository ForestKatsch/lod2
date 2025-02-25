package auth

import (
	"database/sql"
	"log"
	"time"

	"lod2/db"

	"go.jetify.com/typeid"
)

func migrateInvites() {
	db.MigrateTable("authInvites", migrateInvitesTable)
}

// authInvites table has rows:
// inviteId TEXT -- the unique invite code
// userId TEXT -- the user who created this invite
// issuedAt INTEGER - when the invite was created
// inviteLimit INTEGER - how many times this invite can be used

func migrateInvitesTable(tx *sql.Tx, version int) (int, error) {
	// 1: Table exists
	if version < 1 {
		_, err := tx.Exec("CREATE TABLE authInvites (inviteId TEXT PRIMARY KEY NOT NULL UNIQUE, userId TEXT NOT NULL UNIQUE, issuedAt INTEGER NOT NULL, inviteLimit INTEGER NOT NULL) WITHOUT ROWID")

		if err != nil {
			return version, err
		}

		version = 1
	}

	return version, nil
}

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
func adminSetRemainingInvites(userId string, remainingInvites int) error {
	currentRemainingInvites, err := invitesRemaining(userId)

	if err != nil {
		log.Println("error selecting remaining invites:", err)
		return err
	}

	newInviteLimit := remainingInvites - currentRemainingInvites

	_, err = db.DB.Exec("UPDATE authInvites SET inviteLimit = ? WHERE userId = ?", newInviteLimit, userId)

	if err != nil {
		log.Println("error updating invite:", err)
		return err
	}

	return nil
}

func invitesRemaining(userId string) (int, error) {
	var invitesRemaining int

	err := db.DB.QueryRow("SELECT SUM(inviteLimit) - COUNT(*) as invitesRemaining FROM authInvites WHERE userId = ?", userId).Scan(&invitesRemaining)

	if err != nil {
		log.Println("error selecting remaining invites:", err)
		return 0, err
	}

	return invitesRemaining, nil
}
