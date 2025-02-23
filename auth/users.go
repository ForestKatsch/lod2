package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"lod2/db"

	"github.com/mattn/go-sqlite3"
	"go.jetify.com/typeid"
)

func migrateUsers() {
	db.MigrateTable("authUsers", migrateAuthUsersTable)
	//migrateTable(db, "authInvites", migrateAuthUsersTable)
}

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
		userId, err := createUser("admin", "admin")

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
func createUser(username string, password string) (string, error) {
	userId, _ := typeid.WithPrefix("user")
	passwordHash, err := hashPassword(password)

	if err != nil {
		return "", err
	}

	fmt.Printf("userId: %v, username: %v, passwordHash: %v\n", userId, username, passwordHash)
	_, err = db.DB.Exec("INSERT INTO authUsers (userId, userName, userPasswordHash) VALUES (?, ?, ?)", userId, username, passwordHash)

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

// Returns the user ID, or an error if the user does not exist or the password is incorrect.
func getUserLogin(username string, password string) (string, error) {
	var userId string
	var passwordHash string

	err := db.DB.QueryRow("SELECT userId, userPasswordHash FROM authUsers WHERE userName = ?", username).Scan(&userId, &passwordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("invalid username")
		}

		return "", err
	}

	passwordValid := verifyPassword(passwordHash, password)

	if !passwordValid {
		return "", errors.New("invalid password")
	}

	return string(userId), nil
}

// Returns an error if the userId does not exist or their password is incorrect.
func verifyUserPassword(userId string, password string) error {
	var passwordHash string

	err := db.DB.QueryRow("SELECT userPasswordHash FROM authUsers WHERE userId = ?", userId).Scan(&passwordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("invalid user id")
		}

		return err
	}

	passwordValid := verifyPassword(passwordHash, password)

	if !passwordValid {
		return errors.New("invalid password")
	}

	return nil
}

// User-facing change password function. Used by a user to change their own password.
func ChangePassword(userId string, currentPassword string, newPassword string, newPasswordVerify string) error {
	err := verifyUserPassword(userId, currentPassword)

	if err != nil {
		return errors.New("invalid current password")
	}

	if newPassword != newPasswordVerify {
		return errors.New("new passwords do not match")
	}

	newPasswordHash, err := hashPassword(newPassword)

	if err != nil {
		return err
	}

	_, err = db.DB.Exec("UPDATE authUsers SET userPasswordHash = ? WHERE userId = ?", newPasswordHash, userId)

	if err != nil {
		return err
	}

	return nil
}

func AdminGetAllUsers() ([]UserInfo, error) {
	rows, err := db.DB.Query("SELECT userId, userName FROM authUsers")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []UserInfo

	for rows.Next() {
		var userId string
		var userName string

		err := rows.Scan(&userId, &userName)

		if err != nil {
			return nil, err
		}

		users = append(users, UserInfo{UserId: userId, Username: userName})
	}

	return users, nil
}
