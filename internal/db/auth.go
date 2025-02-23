package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"lod2/internal/db/auth"

	"github.com/mattn/go-sqlite3"
	"go.jetify.com/typeid"
)

// authUsersTable has rows:
// userId TEXT
// userName TEXT
// userPasswordHash TEXT
// invitesAvailable INTEGER

func migrateAuthUsersTable(db *sql.DB, version int) (int, error) {
	// 1: Table exists
	if version < 1 {
		db.Exec("CREATE TABLE authUsers (userId TEXT PRIMARY KEY NOT NULL UNIQUE, userName TEXT NOT NULL UNIQUE, userPasswordHash TEXT NOT NULL UNIQUE) WITHOUT ROWID")
		version = 1
	}

	// 2: Adds admin user by default
	if version < 2 {
		userId, err := CreateUser("admin", "admin")

		if err != nil {
			log.Println("failed to create admin user:", err)
			return version, err
		}

		log.Printf("created admin user with ID %s", userId)
		version = 2
	}

	// 3:
	if version < 3 {
		// Add the invitesAvailable column and default this to 0 for all users.
		transaction, _ := db.Begin()
		transaction.Exec("ALTER TABLE authUsers ADD COLUMN invitesAvailable INTEGER DEFAULT 0")
		transaction.Exec("UPDATE authUsers SET invitesAvailable = -1 WHERE userName = 'admin'")
		transaction.Commit()

		version = 3
	}

	return version, nil
}

// Creates a user with the provided username and password.
func CreateUser(username string, password string) (string, error) {
	userId, _ := typeid.WithPrefix("user")
	passwordHash, err := auth.HashPassword(password)

	if err != nil {
		return "", err
	}

	fmt.Printf("userId: %v, username: %v, passwordHash: %v\n", userId, username, passwordHash)
	_, err = db.Exec("INSERT INTO authUsers (userId, userName, userPasswordHash) VALUES (?, ?, ?)", userId, username, passwordHash)

	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return "", errors.New("this username is already taken")
			}
		}

		return "", err
	}

	return userId.String(), nil
}

func GetAllUser

// Returns the user ID, or an error if the user does not exist or the password is incorrect.
func GetUserLogin(username string, password string) (string, error) {
	var userId string
	var passwordHash string

	err := db.QueryRow("SELECT userId, userPasswordHash FROM authUsers WHERE userName = ?", username).Scan(&userId, &passwordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("invalid username")
		}

		return "", err
	}

	passwordValid := auth.VerifyPassword(passwordHash, password)

	if !passwordValid {
		return "", errors.New("invalid password")
	}

	return string(userId), nil
}
