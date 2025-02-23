package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"lod2/db"

	"github.com/mattn/go-sqlite3"
	"go.jetify.com/typeid"
)

func migrateUsers() {
	db.MigrateTable("authUsers", migrateAuthUsersTable)
}

// authUsers has rows:
// userId TEXT
// userName TEXT
// userPasswordHash TEXT
// invitesAvailable INTEGER

func migrateAuthUsersTable(tx *sql.Tx, version int) (int, error) {
	// 1: Table exists
	if version < 1 {
		tx.Exec("CREATE TABLE authUsers (userId TEXT PRIMARY KEY NOT NULL UNIQUE, userName TEXT NOT NULL UNIQUE, userPasswordHash TEXT NOT NULL UNIQUE) WITHOUT ROWID")
		version = 1
	}

	// 2: Adds admin user by default
	if version < 2 {
		userId, _ := createUser(tx, "admin", "admin")
		log.Printf("creating admin user with ID %s", userId)
		version = 2
	}

	// 3:
	if version < 3 {
		// Add the invitesAvailable column and default this to 0 for all users.
		tx.Exec("ALTER TABLE authUsers ADD COLUMN invitesAvailable INTEGER DEFAULT 0")
		tx.Exec("UPDATE authUsers SET invitesAvailable = -1 WHERE userName = 'admin'")

		version = 3
	}

	return version, nil
}

// Creates a user with the provided username and password.
func createUser(tx *sql.Tx, username string, password string) (string, error) {
	userId, _ := typeid.WithPrefix("user")
	passwordHash, err := hashPassword(password)

	if err != nil {
		return "", err
	}

	fmt.Printf("userId: %v, username: %v, passwordHash: %v\n", userId, username, passwordHash)
	_, err = tx.Exec("INSERT INTO authUsers (userId, userName, userPasswordHash) VALUES (?, ?, ?)", userId, username, passwordHash)

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

type UserSessionInfo struct {
	UserId       string
	Username     string
	LastLogin    time.Time
	SessionCount int
}

func AdminGetAllUsers() ([]UserSessionInfo, error) {
	rows, err := db.DB.Query(`
		SELECT authUsers.userId, authUsers.userName, max(authSessions.issuedAt) as lastLogin, count(authSessions.expiresAt < ?) as sessionCount
		FROM authUsers
		LEFT JOIN authSessions ON authUsers.userId = authSessions.userId`, time.Now().Unix())

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []UserSessionInfo

	for rows.Next() {
		var userId string
		var userName string
		var lastLogin int64
		var sessionCount int

		err := rows.Scan(&userId, &userName, &lastLogin, &sessionCount)

		if err != nil {
			return nil, err
		}

		users = append(users, UserSessionInfo{
			UserId:       userId,
			Username:     userName,
			LastLogin:    time.Unix(lastLogin, 0),
			SessionCount: sessionCount,
		})
	}

	return users, nil
}
