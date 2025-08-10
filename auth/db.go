package auth

import (
	"lod2/db"
	"log"
)

// PostMigrationSetup handles admin user setup after database migrations are complete
// This ensures the admin user has proper password hashing and all current roles
func PostMigrationSetup() {
	// Update admin user to have proper password hash and all roles
	// This runs at every boot and makes sure the admin user has all roles even if additional scopes are added.

	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("failed to begin post-migration setup transaction: %v", err)
		return
	}
	defer tx.Rollback()

	userId, err := AdminGetUserIdByUsername(tx, "admin")
	if err != nil {
		// Admin user doesn't exist, create it
		userId, err = createUser(tx, "admin", "admin", AllRoles)
		if err != nil {
			log.Printf("failed to create admin user: %v", err)
			return
		}
		log.Printf("created admin user with ID %s", userId)
	} else {
		// Admin user exists, update roles
		if err := setRoles(tx, userId, AllRoles); err != nil {
			log.Printf("failed to update admin roles: %v", err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("failed to commit post-migration setup: %v", err)
		return
	}
}
